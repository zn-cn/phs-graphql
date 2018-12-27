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

	OwnGroups    []string `bson:"ownGroups" json:"ownGroups"`
	ManageGroups []string `bson:"manageGroups" json:"manageGroups"`
	JoinGroups   []string `bson:"joinGroups" json:"joinGroups"`
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
		"unionid":   1,
		"nickname":  1,
		"avatarUrl": 1,
	}

	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableUser)

	err := table.Find(query).Select(selector).One(&user)
	status := constant.UserFollowStatus
	if ok, err := IsFollowOfficeAccount(userInfo.UnionID); !ok || err != nil {
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
		update := bson.M{
			"$set": bson.M{
				"status":    status,
				"openid":    userInfo.OpenID,
				"nickname":  userInfo.NickName,
				"avatarUrl": userInfo.AvatarURL,
				"gender":    userInfo.Gender,
				"language":  userInfo.Language,
				"country":   userInfo.Country,
				"city":      userInfo.City,
				"province":  userInfo.Province,
			},
		}
		return table.Update(query, update)
	}
	return nil
}

func CreateUserByUnionid(unionid string) error {
	query := bson.M{
		"unionid": unionid,
	}

	selector := bson.M{
		"status": 1,
	}
	oldUser, err := findUser(query, selector)
	if err != nil {
		user := User{
			ID:      bson.NewObjectId(),
			Status:  constant.UserFollowStatus,
			Unionid: unionid,
		}
		return insertUsers(user)
	}
	if oldUser.Status != constant.UserFollowStatus {
		update := bson.M{
			"$set": bson.M{
				"status": constant.UserFollowStatus,
			},
		}
		return updateUser(query, update)
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

func FindGroupsByUserID(unionid string) (ownGroups, manageGroups, joinGroups []string, err error) {
	if unionid == "" {
		err = constant.ErrorIDFormatWrong
		return
	}

	query := bson.M{
		"unionid": unionid,
	}
	selector := bson.M{
		"ownGroups":    1,
		"manageGroups": 1,
		"joinGroups":   1,
	}

	user, err := findUser(query, selector)
	ownGroups, manageGroups, joinGroups = user.OwnGroups, user.ManageGroups, user.JoinGroups
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
			"ownGroups": id,
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

func GetRedisUserInfo(unionid string) (map[string]string, error) {
	ids := []string{unionid}
	userInfos, err := GetRedisUserInfos(ids)
	if len(userInfos) > 0 {
		return userInfos[0], err
	}
	return nil, err
}

func GetRedisUserInfos(unionids []string) ([]map[string]string, error) {
	redisConn := db.NewRedisDBCntlr()
	defer redisConn.Close()

	resData := make([]map[string]string, len(unionids))
	for i, unionid := range unionids {
		key := fmt.Sprintf(constant.RedisUserInfo, unionid)
		userInfo, err := redisConn.HGETALL(key)
		if len(userInfo) == 0 || err != nil {
			user, err := setRedisUserInfo(unionid)
			if err != nil {
				return nil, err
			}
			userInfo = map[string]string{
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
	if err != nil {
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
