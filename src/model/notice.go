package model

import (
	"constant"
	"fmt"
	"model/db"
	"time"
	"util"

	"gopkg.in/mgo.v2/bson"
)

type Notice struct {
	ID     bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Type   int           `bson:"type" json:"type"`     // 类型：1表示前一天发送通知，2表示前两天发送通知
	Status int           `bson:"status" json:"status"` // 状态: -10 删除状态, -1表示过期状态，5表示发布状态

	CreatorID string `bson:"creatorID" json:"creatorID"` // unionid
	GroupID   string `bson:"groupID" json:"groupID"`     // _id

	Title      string `bson:"title" json:"title"` // 标题
	Content    string `bson:"content" json:"content"`
	Imgs       []Img  `bson:"imgs" json:"imgs"`
	Note       string `bson:"note" json:"note"`             // 备注
	CreateTime int64  `bson:"createTime" json:"createTime"` // 创建时间
	NoticeTime int64  `bson:"noticeTime" json:"noticeTime"` // 提醒时间

	WatchUserIDs []string `bson:"watchUserIDs" json:"watchUserIDs"` // 查看用户
	WatchNum     int      `bson:"watchNum" json:"watchNum"`         // 查看人数
	LikeUserIDs  []string `bson:"likeUserIDs" json:"likeUserIDs"`   // 点赞用户
	LikeNum      int      `bson:"likeNum" json:"likeNum"`           // 点赞人数
}

func CreateNotices(userID string, notices []Notice) error {
	docs := make([]interface{}, len(notices))
	now := util.GetNowTimestamp()
	for i := 0; i < len(notices); i++ {
		if notices[i].Title == "" || notices[i].NoticeTime <= now || notices[i].GroupID == "" {
			return constant.ErrorParamWrong
		}
		notices[i].ID = bson.NewObjectId()
		notices[i].Status = constant.NoticePubStatus
		notices[i].CreatorID = userID
		notices[i].CreateTime = now
		docs[i] = notices[i]
	}
	err := insertNotices(docs...)
	if err != nil {
		return err
	}
	go setRedisUserWeekNotice(notices)
	return nil
}

func setRedisUserWeekNotice(notices []Notice) error {
	mgoCntrl := db.NewCopyMgoDBCntlr()
	defer mgoCntrl.Close()
	groupTable := mgoCntrl.GetTable(constant.TableGroup)

	redisCntrl := db.NewRedisDBCntlr()
	defer redisCntrl.Close()

	selector := bson.M{
		"ownerID":    1,
		"managerIDs": 1,
		"memberIDs":  1,
	}

	for _, notice := range notices {
		group := Group{}
		err := groupTable.FindId(bson.ObjectIdHex(notice.GroupID)).Select(selector).One(&group)
		if err != nil {
			break
		}

		memberIDs := append(append([]string{group.OwnerID}, group.ManagerIDs...), group.MemberIDs...)

		t := time.Unix(notice.NoticeTime/1000, 0)
		now := util.GetNowTimestamp()
		weekStart := util.GetWeekStartTimestamp(t)
		weekEnd := util.GetWeekEndTimestamp(t)
		expire := (weekEnd - now) / 1000

		for _, member := range memberIDs {
			key := fmt.Sprintf(constant.RedisUserWeekNotice, weekStart, member)
			redisCntrl.RPUSH(key, notice.ID.Hex())
			redisCntrl.EXPIRE(key, expire)
		}
	}
	return nil
}

func GetNotice(id string) (Notice, error) {
	if !bson.IsObjectIdHex(id) {
		return Notice{}, constant.ErrorIDFormatWrong
	}
	query := bson.M{
		"_id": bson.ObjectIdHex(id),
	}
	return findNotice(query, DefaultSelector)
}

func GetNotices(groups []string, page, perPage int) ([]Notice, error) {
	query := bson.M{
		"groupID": bson.M{
			"$in": groups,
		},
		"status": bson.M{
			"$gte": constant.NoticeExpireStatus,
		},
	}
	selector := bson.M{
		"watchUserIDs": 0,
		"likeUserIDs":  0,
	}
	fields := []string{
		"-status",
		"noticeTime",
	}
	notices, err := findNotices(query, selector, page, perPage, fields...)
	if err != nil {
		return notices, err
	}

	now := util.GetNowTimestamp()


	right := len(notices) - 1
	for i, notice := range notices {
		if notice.NoticeTime < now {
			// 已经过期
			for j := i; j < len(notices) && j < right; j++ {
				notices[j], notices[right] = notices[right], notices[j]
				right--
			}
			break
		}
	}
	return newNotices, err
}

func UpdateNotice(noticeID, userID string, updateData map[string]interface{}) error {
	if !bson.IsObjectIdHex(noticeID) {
		return constant.ErrorIDFormatWrong
	}
	query := bson.M{
		"_id":       bson.ObjectIdHex(noticeID),
		"creatorID": userID,
		"status": bson.M{
			"$gte": constant.NoticePubStatus,
		},
	}
	update := bson.M{
		"$set": updateData,
	}
	return updateNotice(query, update)
}

func SendDayNotice() error {
	now := util.GetNowTimestamp()
	nextDayEnd := util.GetNextDayEndTimestamp()

	cntrl := db.NewCopyMgoDBCntlr()
	defer cntrl.Close()
	noticeTable := cntrl.GetTable(constant.TableNotice)
	query := bson.M{
		"noticeTime": bson.M{
			"$gt": now,
			"$lt": nextDayEnd,
		},
	}

	selector := bson.M{
		"watchUserIDs": 0,
		"watchNum":     0,
		"likeUserIDs":  0,
		"likeNum":      0,
	}
	notices := []Notice{}
	noticeTable.Find(query).Select(selector).All(&notices)
	if len(notices) == 0 {
		return nil
	}

	groupTable := cntrl.GetTable(constant.TableGroup)
	userTable := cntrl.GetTable(constant.TableUser)

	templates := []WechatTemplate{}
	for _, notice := range notices {
		selector = bson.M{
			"ownerID":    1,
			"managerIDs": 1,
			"memberIDs":  1,
		}
		group := Group{}
		err := groupTable.FindId(bson.ObjectIdHex(notice.GroupID)).Select(selector).One(&group)
		if err != nil {
			break
		}

		memberIDs := append(append([]string{group.OwnerID}, group.ManagerIDs...), group.MemberIDs...)
		query = bson.M{
			"unionid": bson.M{
				"$in": memberIDs,
			},
			"status": bson.M{
				"$gte": constant.UserFollowStatus,
			},
		}
		selector = bson.M{
			"unionid":  1,
			"nickname": 1,
		}
		users := []User{}
		userTable.Find(query).Select(selector).All(&users)

		for _, user := range users {
			year, month, day := time.Unix(notice.NoticeTime/1000, 0).Date()
			timeStr := fmt.Sprintf(constant.TemplateTime, year, month, day)
			template := getNoticeTemplate(user.Unionid, notice.Title, notice.Content, timeStr)
			templates = append(templates, template)
		}
	}

	_, err := sendOfficeAccountTemplate(templates)
	return err
}

func SendWeekNotice() error {
	now := time.Now().AddDate(0, 0, 7)
	timestamp := util.GetWeekStartTimestamp(now)
	p := fmt.Sprintf(constant.RedisUserWeek, timestamp)

	redisCntrl := db.NewRedisDBCntlr()
	defer redisCntrl.Close()

	mgoCntrl := db.NewCopyMgoDBCntlr()
	defer mgoCntrl.Close()
	noticeTable := mgoCntrl.GetTable(constant.TableNotice)

	keys, _ := redisCntrl.KEYS(p)
	if len(keys) == 0 {
		return nil
	}
	templates := []WechatTemplate{}
	prefixLen := len(p) - 1

	selector := bson.M{
		"title":      1,
		"content":    1,
		"noticeTime": 1,
	}

	for _, key := range keys {
		unionid := key[prefixLen:]
		noticeIDs, _ := redisCntrl.LRANGE(key, 0, -1)
		noticeBsonIDs := make([]bson.ObjectId, len(noticeIDs))
		for i, id := range noticeIDs {
			noticeBsonIDs[i] = bson.ObjectIdHex(id)
		}
		query := bson.M{
			"_id": bson.M{
				"$in": noticeBsonIDs,
			},
		}
		notices := []Notice{}
		noticeTable.Find(query).Select(selector).All(&notices)
		timeStr := "下周"
		var title string
		var content string
		for _, notice := range notices {
			title += notice.Title + "\n"
			content += notice.Content + "\n"
		}
		template := getNoticeTemplate(unionid, title, content, timeStr)
		templates = append(templates, template)
	}
	_, err := sendOfficeAccountTemplate(templates)
	return err
}

func UpdateExpireNotice() error {
	now := util.GetNowTimestamp()
	query := bson.M{
		"status": bson.M{
			"$gte": constant.NoticePubStatus,
		},
		"noticeTime": bson.M{
			"$lt": now,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"status": constant.NoticeExpireStatus,
		},
	}
	_, err := updateNotices(query, update)
	return err
}

/****************************************** notice basic action ****************************************/

func findNotice(query, selector interface{}) (Notice, error) {
	data := Notice{}
	cntrl := db.NewCopyMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableNotice)
	err := table.Find(query).Select(selector).One(&data)
	return data, err
}

func findNotices(query, selector interface{}, page, perPage int, fields ...string) ([]Notice, error) {
	data := []Notice{}
	cntrl := db.NewCopyMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableNotice)
	err := table.Find(query).Sort(fields...).Select(selector).Skip((page - 1) * perPage).Limit(perPage).All(&data)
	return data, err
}

func findNoticesByRaw(query, selector interface{}) ([]Notice, error) {
	data := []Notice{}
	cntrl := db.NewCopyMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableNotice)
	err := table.Find(query).Select(selector).All(&data)
	return data, err
}

func insertNotices(docs ...interface{}) error {
	return insertDocs(constant.TableNotice, docs...)
}

func updateNotice(query, update interface{}) error {
	return updateDoc(constant.TableNotice, query, update)
}

func updateNotices(query, update interface{}) (interface{}, error) {
	return updateDocs(constant.TableNotice, query, update)
}
