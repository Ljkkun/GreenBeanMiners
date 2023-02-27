package service

import (
	"errors"
	"fmt"
	"github.com/Ljkkun/GreenBeanMiners/global"
	"github.com/Ljkkun/GreenBeanMiners/model"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

// GetFeedVideosAndAuthorsRedis 获取推送视频以及其作者并返回视频数
func GetFeedVideosAndAuthorsRedis(videoList *[]model.Video, authors *[]model.User, LatestTime int64, MaxNumVideo int) (int, error) {
	// 确保 feed 在 redis 中
	if err := GoFeed(); err != nil {
		return 0, err
	}
	// 初始化查询条件， Offset和 Count用于分页
	op := redis.ZRangeBy{
		Min:    "0",                                                         // 最小分数
		Max:    strconv.FormatFloat(float64(LatestTime-2)/1000, 'f', 3, 64), // 最大分数
		Offset: 0,                                                           // 类似sql的limit, 表示开始偏移量
		Count:  int64(MaxNumVideo),                                          // 一次返回多少数据
	}
	// 获取推送视频ID按逆序返回
	videoIDStrList, err := global.REDIS.ZRevRangeByScore(global.CONTEXT, "feed", &op).Result()
	numVideos := len(videoIDStrList)
	if err != nil || numVideos == 0 {
		return 0, err
	}

	videoIDList := make([]uint64, 0, numVideos)
	for _, videoIDStr := range videoIDStrList {
		videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
		if err != nil {
			continue
		}
		videoIDList = append(videoIDList, videoID)
	}
	if err = GetVideoListByIDsRedis(videoList, videoIDList); err != nil {
		return 0, err
	}
	numVideos = len(*videoList)
	// 批量或者视频作者
	authorIDList := make([]uint64, numVideos)
	for i, video := range *videoList {
		authorIDList[i] = video.AuthorID
	}
	if err = GetUserListByUserIDs(authorIDList, authors); err != nil {
		return 0, err
	}
	return numVideos, nil
}

// PublishVideo 将用户上传的视频信息写入数据库
func PublishVideo(userID uint64, videoID uint64, videoName string, coverName string, title string) error {
	video := model.Video{
		VideoID:   videoID,
		Title:     title,
		PlayName:  videoName,
		CoverName: coverName,
		//FavoriteCount: 0,
		//CommentCount:  0,
		AuthorID:  userID,
		CreatedAt: time.Now(),
	}
	if global.DB.Create(&video).Error != nil {
		return errors.New("video表插入失败")
	}
	keyPublish := fmt.Sprintf(PublishPattern, userID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyPublish).Result()
	if err != nil {
		return err
	}

	if n <= 0 {
		//	keyPublish不存在 查询mysql将用户发布过的视频全部写入缓存中
		var videoList []model.Video
		if err = global.DB.Where("author_id = ?", userID).Find(&videoList).Error; err != nil {
			return err
		}
		var listZ = make([]*redis.Z, 0, len(videoList))
		for _, video_ := range videoList {
			listZ = append(listZ, &redis.Z{Score: float64(video_.CreatedAt.UnixMilli()) / 1000, Member: video_.VideoID})
		}
		return PublishEvent(video, listZ...)
	}
	// keyPublish存在 只添加当前上传的视频
	Z := redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: videoID}
	return PublishEvent(video, &Z)
}

// GetPublishedVideosRedis 获取用户上传的视频列表
func GetPublishedVideosRedis(videoList *[]model.Video, userID uint64) (int, error) {
	keyEmpty := fmt.Sprintf(EmptyPattern, userID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyEmpty).Result()
	if n > 0 || err != nil {
		// 当前用户没有发布过视频
		return 0, err
	}
	keyPublish := fmt.Sprintf(PublishPattern, userID)
	n, err = global.REDIS.Exists(global.CONTEXT, keyPublish).Result()
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		// "publish userid"不存在
		// 因为有序集合插入时需要video的创建时间当做score，所以不能只查主键
		result := global.DB.Where("author_id = ?", userID).Find(videoList)
		numVideos := int(result.RowsAffected)

		if result.Error != nil {
			return 0, err
		}
		if numVideos == 0 {
			return 0, SetUserPublishEmpty(userID)
		}
		var listZ = make([]*redis.Z, 0, numVideos)
		var videoIDList = make([]uint64, 0, numVideos)
		for _, video_ := range *videoList {
			listZ = append(listZ, &redis.Z{Score: float64(video_.CreatedAt.UnixMilli()) / 1000, Member: video_.VideoID})
			videoIDList = append(videoIDList, video_.VideoID)
		}
		// 批量查找favorite_count与comment_count
		favoriteCountList, err := GetFavoriteCountListByVideoIDList(videoIDList)
		if err != nil {
			return 0, err
		}
		var commentCountList []int64
		if err = GetCommentCountListByVideoIDList(videoIDList, &commentCountList); err != nil {
			return 0, err
		}
		for i := range *videoList {
			(*videoList)[i].FavoriteCount = favoriteCountList[i]
			(*videoList)[i].CommentCount = commentCountList[i]
		}
		// 将用户发表过的视频列表写入缓存
		if err = GoPublish(userID, listZ...); err != nil {
			return 0, err
		}

		return numVideos, nil
	}
	// keyPublish存在
	if err = global.REDIS.Expire(global.CONTEXT, keyPublish, global.PUBLISH_EXPIRE).Err(); err != nil {
		return 0, err
	}
	videoIDStrList, err := global.REDIS.ZRevRange(global.CONTEXT, keyPublish, 0, -1).Result()
	numVideos := len(videoIDStrList)
	if err != nil {
		return 0, err
	}
	videoIDList := make([]uint64, 0, numVideos)
	for _, videoIDStr := range videoIDStrList {
		videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
		if err != nil {
			continue
		}
		videoIDList = append(videoIDList, videoID)
	}
	if err = GetVideoListByIDsRedis(videoList, videoIDList); err != nil {
		return 0, err
	}
	numVideos = len(*videoList)
	return numVideos, nil
}

// GetVideoListByIDsRedis 给定视频ID列表得到对应的视频信息
func GetVideoListByIDsRedis(videoList *[]model.Video, videoIDs []uint64) error {
	numVideos := len(videoIDs)
	*videoList = make([]model.Video, 0, numVideos)
	inCache := make([]bool, 0, numVideos)
	notInCacheIDList := make([]uint64, 0, numVideos)
	for _, videoID := range videoIDs {
		keyVideo := fmt.Sprintf(VideoPattern, videoID)
		n, err := global.REDIS.Exists(global.CONTEXT, keyVideo).Result()
		if err != nil {
			return err
		}
		if n <= 0 {
			// 当前视频不在缓存中
			*videoList = append(*videoList, model.Video{})
			inCache = append(inCache, false)
			notInCacheIDList = append(notInCacheIDList, videoID)
			continue
		}
		// video存在
		var video model.Video
		if err = global.REDIS.Expire(global.CONTEXT, keyVideo, global.VIDEO_EXPIRE).Err(); err != nil {
			return err
		}
		if err = global.REDIS.HGetAll(global.CONTEXT, keyVideo).Scan(&video); err != nil {
			return errors.New("GetVideoListByIDsRedis fail")
		}
		video.VideoID = videoID
		timeUnixMilliStr, err := global.REDIS.HGet(global.CONTEXT, keyVideo, "created_at").Result()
		if err != nil {
			continue
		}
		timeUnixMilli, err := strconv.ParseInt(timeUnixMilliStr, 10, 64)
		if err != nil {
			continue
		}
		video.CreatedAt = time.UnixMilli(timeUnixMilli)
		*videoList = append(*videoList, video)
		inCache = append(inCache, true)
	}
	if len(notInCacheIDList) == 0 {
		// 视频全部在缓存中则提前返回
		return nil
	}
	// 批量查找不在redis的video
	var notInCacheVideoList []model.Video
	if err := GetVideoListByIDsSql(&notInCacheVideoList, notInCacheIDList); err != nil {
		return err
	}
	// 将不在redis中的video填入返回值
	idxNotInCache := 0
	for i := range *videoList {
		if inCache[i] == false {
			(*videoList)[i] = notInCacheVideoList[idxNotInCache]
			idxNotInCache++
		}
	}
	return nil
}

// GetVideoListByIDsSql 被调用当videoID不在redis中，我们不得不查sql
func GetVideoListByIDsSql(videoList *[]model.Video, videoIDs []uint64) error {
	var uniqueVideoList []model.Video
	result := global.DB.Where("video_id in ?", videoIDs).Find(&uniqueVideoList)
	if result.Error != nil {
		return result.Error
	}
	numVideos := result.RowsAffected
	// 针对查询结果建立映射关系
	*videoList = make([]model.Video, 0, numVideos)
	mapVideoIDToVideo := make(map[uint64]model.Video, numVideos)
	for _, video := range uniqueVideoList {
		mapVideoIDToVideo[video.VideoID] = video
	}
	// 查询favorite_count与comment_count
	var commentCountList []int64
	if err := GetCommentCountListByVideoIDListSql(videoIDs, &commentCountList); err != nil {
		return err
	}
	favoriteCountList, err := GetFavoriteCountListByVideoIDList(videoIDs)
	if err != nil {
		return err
	}
	for i, videoID := range videoIDs {
		tmpVideo := mapVideoIDToVideo[videoID]
		tmpVideo.FavoriteCount = favoriteCountList[i]
		tmpVideo.CommentCount = commentCountList[i]
		*videoList = append(*videoList, tmpVideo)
	}
	// 当视频信息写入缓存
	return GoVideoList(*videoList)
}

// GetVideoIDListByUserID 得到用户发表过的视频id列表
func GetVideoIDListByUserID(userID uint64, videoIDList *[]uint64) error {
	keyEmpty := fmt.Sprintf(EmptyPattern, userID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyEmpty).Result()
	if n > 0 || err != nil {
		// 当前用户没有发布过视频
		return err
	}
	keyPublish := fmt.Sprintf(VideoCommentsPattern, userID)
	n, err = global.REDIS.Exists(global.CONTEXT, keyPublish).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// "publish userid"不存在
		var videoList []model.Video
		result := global.DB.Where("author_id = ?", userID).Find(&videoList)
		if result.Error != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return nil
		}
		numVideos := int(result.RowsAffected)
		*videoIDList = make([]uint64, numVideos)
		var listZ = make([]*redis.Z, 0, numVideos)
		for i, videoID := range videoList {
			// 逆序 最新的放在前面
			(*videoIDList)[numVideos-i-1] = videoID.VideoID
		}
		for _, video := range videoList {
			listZ = append(listZ, &redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: video.VideoID})
		}
		// 写入缓存
		return GoPublish(userID, listZ...)
	}
	// "publish userid"存在
	if err = global.REDIS.Expire(global.CONTEXT, keyPublish, global.PUBLISH_EXPIRE).Err(); err != nil {
		return err
	}
	// 逆序 最新的放在前面
	videoIDStrList, err := global.REDIS.ZRevRange(global.CONTEXT, keyPublish, 0, -1).Result()
	numVideos := len(videoIDStrList)
	*videoIDList = make([]uint64, 0, numVideos)
	for _, videoIDStr := range videoIDStrList {
		videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
		if err != nil {
			continue
		}
		*videoIDList = append(*videoIDList, videoID)
	}
	return nil
}
