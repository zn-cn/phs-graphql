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

	AvatarURL  string   `bson:"avatar_url" json:"avatar_url"`   // 群头像
	Nickname   string   `bson:"nickname" json:"nickname"`       // 圈子昵称
	OwnerID    string   `bson:"owner_id" json:"owner_id"`       // unionid 注：以下三种身份不会重复，如：members中不会有owner
	CreateTime int64    `bson:"create_time" json:"create_time"` // 创建时间
	Managers   []string `bson:"managers" json:"managers"`       // 管理员
	Members    []string `bson:"members" json:"members"`         // 成员
	PersonNum  int      `bson:"person_num" json:"person_num"`   // 总人数：1 + 管理员人数 + 成员人数
}

var (
	groupCodeNextNumMutex sync.Mutex
)

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

func GetGroup(id string) (Group, error) {
	query := bson.M{
		"_id": bson.ObjectIdHex(id),
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
		"owner_id": bson.M{
			"$ne": unionid,
		},
		"managers": bson.M{
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
				"members": unionid,
			},
			"$inc": bson.M{
				"person_num": 1,
			},
		}
	} else {
		update = bson.M{
			"$pull": bson.M{
				"members": unionid,
			},
			"$inc": bson.M{
				"person_num": -1,
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
				"join_groups": group.ID.Hex(),
			},
		}
	} else {
		update = bson.M{
			"$pull": bson.M{
				"join_groups": group.ID.Hex(),
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
		"_id":      bson.ObjectIdHex(groupID),
		"owner_id": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"managers": toUserIDs[0],
	}
	update := bson.M{
		"$set": bson.M{
			"owner_id": toUserIDs[0],
		},
		"$pull": bson.M{
			"managers": toUserIDs[0],
		},
		"$addToSet": bson.M{
			"members": ownerID,
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
			"own_groups": groupID,
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
			"own_groups": groupID,
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
		"_id":      bson.ObjectIdHex(groupID),
		"owner_id": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
	}

	group := Group{}
	selector := bson.M{
		"owner_id": 1,
		"managers": 1,
		"members":  1,
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
			"own_groups": groupID,
		},
	}
	userTable.Update(query, update)

	query = bson.M{
		"unionid": bson.M{
			"$in": group.Managers,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"manage_groups": groupID,
		},
	}
	userTable.UpdateAll(query, update)

	query = bson.M{
		"unionid": bson.M{
			"$in": group.Members,
		},
	}
	update = bson.M{
		"$pull": bson.M{
			"join_groups": groupID,
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
		"_id":      bson.ObjectIdHex(groupID),
		"owner_id": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"members": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$addToSet": bson.M{
			"managers": bson.M{
				"$each": toUserIDs,
			},
		},
		"$pullAll": bson.M{
			"members": toUserIDs,
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
			"manage_groups": groupID,
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
			"join_groups": groupID,
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
		"_id":      bson.ObjectIdHex(groupID),
		"owner_id": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"managers": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$pullAll": bson.M{
			"managers": toUserIDs,
		},
		"$addToSet": bson.M{
			"members": bson.M{
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
			"manage_groups": groupID,
		},
		"$addToSet": bson.M{
			"join_groups": groupID,
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
		"_id":      bson.ObjectIdHex(groupID),
		"owner_id": ownerID,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
		"managers": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$pullAll": bson.M{
			"managers": toUserIDs,
		},
		"$inc": bson.M{
			"person_num": -len(toUserIDs),
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
			"manage_groups": groupID,
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
				"owner_id": userID,
			},
			bson.M{
				"managers": userID,
			},
		},
		"members": bson.M{
			"$all": toUserIDs,
		},
	}
	update := bson.M{
		"$pullAll": bson.M{
			"members": toUserIDs,
		},
		"$inc": bson.M{
			"person_num": -len(toUserIDs),
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
			"join_groups": groupID,
		},
	}
	userTable := cntrl.GetTable(constant.TableUser)
	_, err = userTable.UpdateAll(query, update)
	return err
}

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
		cntrl.SET(constant.RedisGroupCodeNextNum, nextNum)
		groupCodeNextNumMutex.Unlock()
	}

	return cntrl.LPOP(constant.RedisGroupCodePool)
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
			data["avatar_url"] = group.AvatarURL
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
		"nickname":   1,
		"avatar_url": 1,
		"code":       1,
	}
	group, err = findGroup(query, selector)
	if err != nil {
		return
	}
	key := fmt.Sprint(constant.RedisGroupInfo, group.ID.Hex())
	args := []interface{}{
		"nickname",
		group.Nickname,
		"avatar_url",
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
