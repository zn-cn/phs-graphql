package model

/*
   圈子群体
*/
import (
	"constant"
	"errors"
	"fmt"
	"model/db"
	"sync"
	"util"

	"gopkg.in/mgo.v2/bson"
)

type Group struct {
	ID bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	// 状态: -10 表示解散状态, 5 表示正常状态
	Status int    `bson:"status" json:"status"`
	Code   string `bson:"code" json:"code"` // 圈子code -> 邀请码, unique

	AvatarURL  string   `bson:"avatarUrl" json:"avatarUrl"`   // 群头像
	Nickname   string   `bson:"nickname" json:"nickname"`     // 圈子昵称
	OwnerID    string   `bson:"ownerID" json:"ownerID"`       // unionid 注：以下三种身份不会重复，如：memberIDs中不会有owner
	CreateTime int64    `bson:"createTime" json:"createTime"` // 创建时间
	ManagerIDs []string `bson:"managerIDs" json:"managerIDs"` // 管理员
	MemberIDs  []string `bson:"memberIDs" json:"memberIDs"`   // 成员
	PersonNum  int      `bson:"personNum" json:"personNum"`   // 总人数：1 + 管理员人数 + 成员人数
}

var groupCodeNextNumMutex sync.Mutex

func init() {
	InitGroupCodeNextNum()
}

func InitGroupCodeNextNum() {
	cntrl := db.NewRedisDBCntlr()
	defer cntrl.Close()

	nextNum, _ := cntrl.GETInt64(constant.RedisGroupCodeNextNum)
	if nextNum == 0 {
		cntrl.SET(constant.RedisGroupCodeNextNum, constant.RedisGroupInitNextNum)
	}
}

func CreateGroup(unionid, nickname, avatarURL string) (string, error) {
	var code string
	for i := 0; i < 5; i++ {
		code, _ = getGroupCode()
		if code != "" {
			break
		}
	}

	group := Group{
		ID:         bson.NewObjectId(),
		Status:     constant.GroupCommonStatus,
		CreateTime: util.GetNowTimestamp(),
		Code:       code,
		AvatarURL:  avatarURL,
		Nickname:   nickname,
		OwnerID:    unionid,
		PersonNum:  1,
	}
	err := insertGroups(group)
	if err != nil {
		return code, err
	}

	go AddUserOwnGroup(unionid, group.ID.Hex())
	return code, err
}

func GetGroupByCode(code string) (Group, error) {
	query := bson.M{
		"code": code,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
	}
	return findGroup(query, DefaultSelector)
}

func JoinGroup(code, unionid string) error {
	return groupAction(code, unionid, true)
}

func LeaveGroup(code, unionid string) error {
	return groupAction(code, unionid, false)
}

func groupAction(code, unionid string, isJoin bool) error {
	query := bson.M{
		"code": code,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"ownerID": bson.M{
			"$ne": unionid,
		},
		"managerIDs": bson.M{
			"$nin": []string{unionid},
		},
	}

	group, err := findGroup(query, bson.M{"code": 1})
	if err != nil {
		return err
	}
	var update bson.M
	if isJoin {
		update = bson.M{
			"$addToSet": bson.M{
				"memberIDs": unionid,
			},
			"$inc": bson.M{
				"personNum": 1,
			},
		}
	} else {
		update = bson.M{
			"$pull": bson.M{
				"memberIDs": unionid,
			},
			"$inc": bson.M{
				"personNum": -1,
			},
		}
	}

	err = updateGroup(query, update)
	if err != nil {
		return err
	}
	query = bson.M{
		"unionid": unionid,
	}

	if isJoin {
		update = bson.M{
			"$addToSet": bson.M{
				"joinGroups": group.ID.Hex(),
			},
		}
	} else {
		update = bson.M{
			"$pull": bson.M{
				"joinGroups": group.ID.Hex(),
			},
		}
	}
	return updateUser(query, update)
}

// UpdateGroupOwner 转让群组, toUserIDs 为 转给的人的id, len = 1, 且只能转给管理员
func UpdateGroupOwner(groupID, ownerID string, toUserIDs []string) error {
	if len(toUserIDs) < 1 {
		return constant.ErrorParamWrong
	}

	if !bson.IsObjectIdHex(groupID) {
		return constant.ErrorIDFormatWrong
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	groupTable := cntrl.GetTable(constant.TableGroup)

	query := bson.M{
		"_id":     bson.ObjectIdHex(groupID),
		"ownerID": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"managerIDs": toUserIDs[0],
	}
	update := bson.M{
		"$set": bson.M{
			"ownerID": toUserIDs[0],
		},
		"$pull": bson.M{
			"managerIDs": toUserIDs[0],
		},
		"$addToSet": bson.M{
			"memberIDs": ownerID,
		},
	}

	err := groupTable.Update(query, update)
	if err != nil {
		return err
	}

	query = bson.M{
		"unionid": ownerID,
	}
	update = bson.M{
		"$pull": bson.M{
			"ownGroups": groupID,
		},
	}
	userTable := cntrl.GetTable(constant.TableUser)
	err = userTable.Update(query, update)
	if err != nil {
		return err
	}

	query = bson.M{
		"unionid": toUserIDs[0],
	}
	update = bson.M{
		"$addToSet": bson.M{
			"ownGroups": groupID,
		},
	}
	return userTable.Update(query, update)

}

// DelGroupOwner 解散群组
func DelGroupOwner(groupID, ownerID string, toUserIDs []string) error {
	if !bson.IsObjectIdHex(groupID) {
		return constant.ErrorIDFormatWrong
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	groupTable := cntrl.GetTable(constant.TableGroup)

	query := bson.M{
		"_id":     bson.ObjectIdHex(groupID),
		"ownerID": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
	}

	group := Group{}
	selector := bson.M{
		"ownerID":    1,
		"managerIDs": 1,
		"memberIDs":  1,
	}
	err := groupTable.FindId(bson.ObjectIdHex(groupID)).Select(selector).One(&group)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status": constant.GroupDelStatus,
		},
	}

	err = groupTable.Update(query, update)
	if err != nil {
		return err
	}

	// 创建者、管理员、成员 更新
	userTable := cntrl.GetTable(constant.TableUser)
	query = bson.M{
		"unionid": ownerID,
	}
	update = bson.M{
		"$pull": bson.M{
			"ownGroups": groupID,
		},
	}
	userTable.Update(query, update)

	query = bson.M{
		"unionid": bson.M{
			"$in": group.ManagerIDs,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"manageGroups": groupID,
		},
	}
	userTable.UpdateAll(query, update)

	query = bson.M{
		"unionid": bson.M{
			"$in": group.MemberIDs,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"joinGroups": groupID,
		},
	}
	userTable.UpdateAll(query, update)
	return nil
}

// SetGroupManager 设置群组管理员
func SetGroupManager(groupID, ownerID string, toUserIDs []string) error {
	if !bson.IsObjectIdHex(groupID) {
		return constant.ErrorIDFormatWrong
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	groupTable := cntrl.GetTable(constant.TableGroup)

	query := bson.M{
		"_id":     bson.ObjectIdHex(groupID),
		"ownerID": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"memberIDs": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$addToSet": bson.M{
			"managerIDs": bson.M{
				"$each": toUserIDs,
			},
		},
		"$pullAll": bson.M{
			"memberIDs": toUserIDs,
		},
	}

	err := groupTable.Update(query, update)
	if err != nil {
		return err
	}

	userTable := cntrl.GetTable(constant.TableUser)
	query = bson.M{
		"unionid": bson.M{
			"$in": toUserIDs,
		},
	}
	update = bson.M{
		"$addToSet": bson.M{
			"manageGroups": groupID,
		},
	}
	userTable.UpdateAll(query, update)

	query = bson.M{
		"unionid": bson.M{
			"$in": toUserIDs,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"joinGroups": groupID,
		},
	}
	_, err = userTable.UpdateAll(query, update)
	return err
}

// UnSetGroupManager 取消群组管理员权限
func UnSetGroupManager(groupID, ownerID string, toUserIDs []string) error {
	if !bson.IsObjectIdHex(groupID) {
		return constant.ErrorIDFormatWrong
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	groupTable := cntrl.GetTable(constant.TableGroup)

	query := bson.M{
		"_id":     bson.ObjectIdHex(groupID),
		"ownerID": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"managerIDs": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$pullAll": bson.M{
			"managerIDs": toUserIDs,
		},
		"$addToSet": bson.M{
			"memberIDs": bson.M{
				"$each": toUserIDs,
			},
		},
	}

	err := groupTable.Update(query, update)
	if err != nil {
		return err
	}

	query = bson.M{
		"unionid": bson.M{
			"$in": toUserIDs,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"manageGroups": groupID,
		},
		"$addToSet": bson.M{
			"joinGroups": groupID,
		},
	}
	userTable := cntrl.GetTable(constant.TableUser)
	_, err = userTable.UpdateAll(query, update)
	return err
}

// DelGroupManager 删除群组管理员
func DelGroupManager(groupID, ownerID string, toUserIDs []string) error {
	if !bson.IsObjectIdHex(groupID) {
		return constant.ErrorIDFormatWrong
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	groupTable := cntrl.GetTable(constant.TableGroup)

	query := bson.M{
		"_id":     bson.ObjectIdHex(groupID),
		"ownerID": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"managerIDs": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$pullAll": bson.M{
			"managerIDs": toUserIDs,
		},
		"$inc": bson.M{
			"personNum": -len(toUserIDs),
		},
	}

	err := groupTable.Update(query, update)
	if err != nil {
		return err
	}

	query = bson.M{
		"unionid": bson.M{
			"$in": toUserIDs,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"manageGroups": groupID,
		},
	}
	userTable := cntrl.GetTable(constant.TableUser)
	_, err = userTable.UpdateAll(query, update)
	return err
}

// DelGroupMember 删除群组成员, 管理员和创建者均可删除成员
func DelGroupMember(groupID, userID string, toUserIDs []string) error {
	if !bson.IsObjectIdHex(groupID) {
		return constant.ErrorIDFormatWrong
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	groupTable := cntrl.GetTable(constant.TableGroup)

	query := bson.M{
		"_id": bson.ObjectIdHex(groupID),
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"$or": []bson.M{
			bson.M{
				"ownerID": userID,
			},
			bson.M{
				"managerIDs": userID,
			},
		},
		"memberIDs": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$pullAll": bson.M{
			"memberIDs": toUserIDs,
		},
		"$inc": bson.M{
			"personNum": -len(toUserIDs),
		},
	}

	err := groupTable.Update(query, update)
	if err != nil {
		return err
	}

	query = bson.M{
		"unionid": bson.M{
			"$in": toUserIDs,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"joinGroups": groupID,
		},
	}
	userTable := cntrl.GetTable(constant.TableUser)
	_, err = userTable.UpdateAll(query, update)
	return err
}

/****************************************** group basic action ****************************************/

func findGroup(query, selectField interface{}) (Group, error) {
	data := Group{}
	cntrl := db.NewCopyMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableGroup)
	err := table.Find(query).Select(selectField).One(&data)
	return data, err
}

func updateGroup(query, update interface{}) error {
	return updateDoc(constant.TableGroup, query, update)
}

func insertGroups(docs ...interface{}) error {
	return insertDocs(constant.TableGroup, docs...)
}

/****************************************** group redis action ****************************************/

func getGroupCode() (string, error) {
	cntrl := db.NewRedisDBCntlr()
	defer cntrl.Close()

	len, _ := cntrl.LLEN(constant.RedisGroupCodePool)
	if len == 0 {
		groupCodeNextNumMutex.Lock()
		nextNum, _ := cntrl.GETInt64(constant.RedisGroupCodeNextNum)
		if nextNum == 0 {
			return "", errors.New("next_num wrong")
		}
		codePool := make([]interface{}, constant.RedisGroupCodePoolNum)
		for i := 0; i < constant.RedisGroupCodePoolNum; i++ {
			nextNum++
			codePool[i] = string(util.Base34(uint64(nextNum)))
		}
		cntrl.RPUSH(constant.RedisGroupCodePool, codePool...)
		cntrl.INCRBY(constant.RedisGroupCodeNextNum, constant.RedisGroupCodePoolNum)
		groupCodeNextNumMutex.Unlock()
	}

	return cntrl.LPOP(constant.RedisGroupCodePool)
}

func GetRedisGroupInfos(ids []string) ([]map[string]string, error) {
	cntrl := db.NewRedisDBCntlr()
	defer cntrl.Close()

	res := make([]map[string]string, len(ids))
	for i, id := range ids {
		key := fmt.Sprint(constant.RedisGroupInfo, id)
		data, err := cntrl.HGETALL(key)
		if len(data) == 0 || err != nil {
			group, err := setRedisGroupInfo(id)
			if err != nil {
				return res, err
			}
			data["nickname"] = group.Nickname
			data["avatarUrl"] = group.AvatarURL
			data["code"] = group.Code
		}
		data["id"] = id
		res[i] = data
	}
	return res, nil
}

func setRedisGroupInfo(id string) (group Group, err error) {
	query := bson.M{
		"_id": bson.ObjectIdHex(id),
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
	}
	selector := bson.M{
		"nickname":  1,
		"avatarUrl": 1,
		"code":      1,
	}
	group, err = findGroup(query, selector)
	if err != nil {
		return
	}
	key := fmt.Sprint(constant.RedisGroupInfo, group.ID.Hex())
	args := []interface{}{
		"nickname",
		group.Nickname,
		"avatarUrl",
		group.AvatarURL,
		"code",
		group.Code,
	}
	cntrl := db.NewRedisDBCntlr()
	defer cntrl.Close()
	_, err = cntrl.HMSET(key, args...)
	cntrl.EXPIRE(key, getRedisDefaultExpire())
	return
}
