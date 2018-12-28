package model

import (
	"constant"
	"fmt"
	"model/db"
	"strconv"
	"util"

	"github.com/imroc/req"

	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	// 状态:  -10表示被删除，0 表示没有关注公众号，5 表示已经关注公众号的普通用户
	Status int `bson:"status" json:"status"`

	// WeixinUserInfo
	Openid    string `bson:"openid" json:"openid"`       // openid
	Unionid   string `bson:"unionid" json:"userID"`      // unionid
	Nickname  string `bson:"nickname" json:"nickname"`   // 用户昵称
	Gender    int    `bson:"gender" json:"gender"`       // 性别 0：未知、1：男、2：女
	Province  string `bson:"province" json:"province"`   // 省份
	City      string `bson:"city" json:"city"`           // 城市
	Country   string `bson:"country" json:"country"`     // 国家
	AvatarURL string `bson:"avatarUrl" json:"avatarUrl"` // 用户头像
	Language  string `bson:"language" json:"language"`   // 语言

	OwnGroupIDs    []string `bson:"ownGroupIDs" json:"ownGroupIDs"`
	ManageGroupIDs []string `bson:"manageGroupIDs" json:"manageGroupIDs"`
	JoinGroupIDs   []string `bson:"joinGroupIDs" json:"joinGroupIDs"`
}

func CreateUser(userInfo *util.DecryptUserInfo) error {
	if userInfo.UnionID == "" {
		return constant.ErrorIDFormatWrong
	}
	if userInfo.AvatarURL == "" {
		userInfo.AvatarURL = constant.WechatDefaultHeadImgURL
	}

	query := bson.M{
		"unionid": userInfo.UnionID,
	}
	user := User{}
	selector := bson.M{
		"unionid":    1,
		"nickname":   1,
		"avatar_url": 1,
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableUser)

	err := table.Find(query).Select(selector).One(&user)
	status := constant.UserFollowStatus
	if ok, _ := IsFollowOfficeAccount(userInfo.UnionID); !ok {
		status = constant.UserUnFollowStatus
	}
	if err != nil {
		user = User{
			ID:        bson.NewObjectId(),
			Status:    status,
			Openid:    userInfo.OpenID,
			Unionid:   userInfo.UnionID,
			Nickname:  userInfo.NickName,
			AvatarURL: userInfo.AvatarURL,
			Gender:    userInfo.Gender,
			Language:  userInfo.Language,
			Country:   userInfo.Country,
			City:      userInfo.City,
			Province:  userInfo.Province,
		}
		return table.Insert(user)
	}

	// update
	if userInfo.NickName != user.Nickname || userInfo.AvatarURL != user.AvatarURL {
		updateMap := map[string]interface{}{
			"status":     status,
			"nickname":   userInfo.NickName,
			"avatar_url": userInfo.AvatarURL,
			"gender":     userInfo.Gender,
			"language":   userInfo.Language,
			"country":    userInfo.Country,
			"city":       userInfo.City,
			"province":   userInfo.Province,
		}
		if userInfo.OpenID != "" {
			updateMap["openid"] = userInfo.OpenID
		}
		update := bson.M{
			"$set": updateMap,
		}
		return table.Update(query, update)
	}
	return nil
}

func GetUserByUnionid(unionid string) (User, error) {
	query := bson.M{
		"unionid": unionid,
	}
	return findUser(query, DefaultSelector)
}

func GetUserStatus(unionid string) (int, error) {
	query := bson.M{
		"unionid": unionid,
	}
	selector := bson.M{
		"status": 1,
	}
	user, err := findUser(query, selector)
	return user.Status, err
}

func SetUserFollowStatus(unionid string, isFollow bool) error {
	oldStatus := constant.UserFollowStatus
	if isFollow {
		oldStatus = constant.UserUnFollowStatus
	}
	query := bson.M{
		"unionid": unionid,
		"status":  oldStatus,
	}
	status := constant.UserFollowStatus
	if !isFollow {
		status = constant.UserUnFollowStatus
	}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}
	return updateUser(query, update)
}

func FindGroupsByUserID(unionid string) (ownGroupIDs, manageGroupIDs, joinGroupIDs []string, err error) {
	if unionid == "" {
		err = constant.ErrorIDFormatWrong
		return
	}

	query := bson.M{
		"unionid": unionid,
	}
	selector := bson.M{
		"ownGroupIDs":    1,
		"manageGroupIDs": 1,
		"joinGroupIDs":   1,
	}

	user, err := findUser(query, selector)
	ownGroupIDs, manageGroupIDs, joinGroupIDs = user.OwnGroupIDs, user.ManageGroupIDs, user.JoinGroupIDs
	return
}

func IsFollowOfficeAccount(unionid string) (bool, error) {
	param := req.Param{
		"unionid": unionid,
	}
	r, _ := req.Get(constant.URLBingYanIsFollow, param)
	resData := struct {
		IsFollow bool `json:"is_follow"`
	}{}
	err := r.ToJSON(&resData)
	return resData.IsFollow, err
}

func AddUserOwnGroup(unionid, id string) error {
	if !bson.IsObjectIdHex(id) {
		return constant.ErrorIDFormatWrong
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableUser)
	query := bson.M{
		"unionid": unionid,
	}
	update := bson.M{
		"$addToSet": bson.M{
			"ownGroupIDs": id,
		},
	}
	return table.Update(query, update)
}

/****************************************** user basic action ****************************************/

func findUser(query, selector interface{}) (User, error) {
	data := User{}
	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableUser)
	err := table.Find(query).Select(selector).One(&data)
	return data, err
}

func updateUser(query, update interface{}) error {
	return updateDoc(constant.TableUser, query, update)
}

func insertUsers(docs ...interface{}) error {
	return insertDocs(constant.TableUser, docs...)
}

/****************************************** user redis action ****************************************/

func GetRedisUserInfo(unionid string) (map[string]interface{}, error) {
	ids := []string{unionid}
	userInfos, err := GetRedisUserInfos(ids)
	if len(userInfos) > 0 {
		return userInfos[0], err
	}
	return nil, err
}

func GetRedisUserInfos(unionids []string) ([]map[string]interface{}, error) {
	redisConn := db.NewRedisDBCntlr()
	defer redisConn.Close()

	resData := make([]map[string]interface{}, len(unionids))
	for i, unionid := range unionids {
		key := fmt.Sprintf(constant.RedisUserInfo, unionid)
		userInfo, err := redisConn.HGETALL(key)
		if len(userInfo) == 0 || err != nil {
			user, _ := setRedisUserInfo(unionid)
			userInfo = map[string]interface{}{
				"nickname":  user.Nickname,
				"gender":    strconv.Itoa(user.Gender),
				"province":  user.Province,
				"city":      user.City,
				"country":   user.Country,
				"avatarUrl": user.AvatarURL,
				"language":  user.Language,
			}
		}
		userInfo["userID"] = unionid
		resData[i] = userInfo
	}
	return resData, nil
}

func setRedisUserInfo(unionid string) (User, error) {
	query := bson.M{
		"unionid": unionid,
	}
	selector := bson.M{
		"nickname":  1,
		"gender":    1,
		"province":  1,
		"city":      1,
		"country":   1,
		"avatarUrl": 1,
		"language":  1,
	}
	user, err := findUser(query, selector)
	if err != nil || user.Nickname == "" || user.AvatarURL == "" {
		return user, err
	}

	cntrl := db.NewRedisDBCntlr()
	defer cntrl.Close()

	key := fmt.Sprintf(constant.RedisUserInfo, unionid)
	args := []interface{}{
		"nickname",
		user.Nickname,
		"gender",
		user.Gender,
		"province",
		user.Province,
		"city",
		user.City,
		"country",
		user.Country,
		"avatarUrl",
		user.AvatarURL,
		"language",
		user.Language,
	}
	cntrl.HMSET(key, args...)
	cntrl.EXPIRE(key, getRedisDefaultExpire())
	return user, nil
}
