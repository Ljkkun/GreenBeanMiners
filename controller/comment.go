package controller

import (
	"github.com/Ljkkun/GreenBeanMiners/global"
	"github.com/Ljkkun/GreenBeanMiners/model"
	"github.com/Ljkkun/GreenBeanMiners/service"
	"github.com/Ljkkun/GreenBeanMiners/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"unicode/utf8"
)

// CommentActionRequest 评论操作的请求
type CommentActionRequest struct {
	UserID      uint64 `form:"user_id" json:"user_id"` // 文档没有传 user_id 这个参数
	Token       string `form:"token" json:"token"`
	VideoID     uint64 `form:"video_id" json:"video_id"`
	ActionType  uint   `form:"action_type" json:"action_type"`
	CommentText string `form:"comment_text" json:"comment_text"`
	CommentID   uint64 `form:"comment_id" json:"comment_id"`
}

type CommentActionResponse struct {
	Response
	Comment Comment `json:"comment,omitempty"`
}

// CommentListRequest 评论列表的请求
type CommentListRequest struct {
	UserID  uint64 `form:"user_id" json:"user_id"`
	Token   string `form:"token" json:"token"`
	VideoID uint64 `form:"video_id" json:"video_id"`
}

// CommentListResponse 评论列表的响应
type CommentListResponse struct {
	Response
	CommentList []Comment `json:"comment_list,omitempty"`
}

// CommentAction 评论操作接口
// 1. 确保操作类型正确 2. 确保当前用户有权限删除
func CommentAction(c *gin.Context) {
	// 参数绑定
	var r CommentActionRequest
	if err := c.ShouldBind(&r); err != nil {
		c.JSON(500, Response{StatusCode: 1, StatusMsg: "bind error"})
		return
	}

	// 判断 action_type 是否正确
	if r.ActionType != 1 && r.ActionType != 2 {
		// action_type 不合法
		c.JSON(400, Response{StatusCode: 1, StatusMsg: "action type error"})
		return
	}

	// 获取 userID
	r.UserID = c.GetUint64("UserID")

	// 评论操作 (发布评论)
	if r.ActionType == 1 {
		// 判断comment是否合法
		if utf8.RuneCountInString(r.CommentText) > global.MAX_COMMENT_LENGTH ||
			utf8.RuneCountInString(r.CommentText) <= 0 {
			c.JSON(200, Response{StatusCode: 1, StatusMsg: "非法评论"})
			return
		}
		// 添加评论
		commentID, err := global.ID_GENERATOR.NextID()
		if err != nil {
			// 生成ID失败
			c.JSON(500, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
		commentModel := model.Comment{
			CommentID: commentID,
			VideoID:   r.VideoID,
			UserID:    r.UserID,
			Content:   r.CommentText,
		}
		// 评论失败
		if err = service.AddComment(&commentModel); err != nil {
			c.JSON(500, Response{StatusCode: 1, StatusMsg: "comment failed"})
			return
		}
		// 未找到评论的用户
		userModel, err := service.UserInfoByUserID(commentModel.UserID)
		if err != nil {
			c.JSON(500, Response{StatusCode: 1, StatusMsg: "comment failed"})
			return
		}
		// 批量判断用户是否关注
		isFollow, err := service.GetFollowStatus(r.UserID, userModel.UserID)
		if err != nil {
			c.JSON(500, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
		// 返回JSON
		c.JSON(http.StatusOK, CommentActionResponse{
			Response: Response{StatusCode: 0},
			Comment: Comment{
				Id: int64(commentModel.CommentID),
				User: User{
					Id:             userModel.UserID,
					Name:           userModel.Name,
					FollowCount:    userModel.FollowCount,
					FollowerCount:  userModel.FollowerCount,
					TotalFavorited: userModel.TotalFavorited,
					FavoriteCount:  userModel.FavoriteCount,
					IsFollow:       isFollow,
				},
				Content:    commentModel.Content,
				CreateDate: commentModel.CreatedAt.Format("2006-01-02 15:04"),
			},
		})
		return
	}

	// 删除评论
	if err := service.DeleteComment(r.UserID, r.VideoID, r.CommentID); err != nil {
		c.JSON(500, Response{StatusCode: 1, StatusMsg: "comment failed"})
		return
	}
	c.JSON(http.StatusOK, Response{StatusCode: 0})
}

// CommentList 评论列表接口
func CommentList(c *gin.Context) {
	// 参数绑定
	var r CommentListRequest
	if err := c.ShouldBind(&r); err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "bind error"})
		return
	}

	var commentModelList []model.Comment
	var userModelList []model.User
	// 获取评论列表以及对应的作者
	if err := service.GetCommentListAndUserListRedis(r.VideoID, &commentModelList, &userModelList); err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}

	var (
		isFollowList []bool
		isLogged     = false // 用户是否传入了合法有效的token（是否登录）
		isFollow     bool
		err          error
	)

	var userID uint64
	// 判断传入的token是否合法，用户是否存在
	if token := c.Query("token"); token != "" {
		claims, err := util.ParseToken(token)
		if err == nil {
			// token合法
			userID = claims.UserID
			isLogged = true
		}
	}

	if isLogged {
		// 当用户登录时 一次性获取用户是否点赞了列表中的视频以及是否关注了评论的作者
		authorIDList := make([]uint64, len(commentModelList))
		for i, user_ := range userModelList {
			authorIDList[i] = user_.UserID
		}
		// 批量判断用户是否关注评论的作者
		isFollowList, err = service.GetFollowStatusList(userID, authorIDList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
	}

	var (
		commentJsonList = make([]Comment, 0, len(commentModelList))
		commentJson     Comment
		userJson        User
		user            model.User
	)

	for i, comment := range commentModelList {
		// 未登录时默认为未关注未点赞
		isFollow = false
		if isLogged {
			// 当用户登录时，判断是否关注当前作者
			isFollow = isFollowList[i]
		}
		user = userModelList[i]
		userJson.Id = user.UserID
		userJson.Name = user.Name
		userJson.FollowCount = user.FollowCount
		userJson.FollowerCount = user.FollowerCount
		userJson.TotalFavorited = user.TotalFavorited
		userJson.FavoriteCount = user.FavoriteCount
		userJson.IsFollow = isFollow

		commentJson.Id = int64(comment.CommentID)
		commentJson.User = userJson
		commentJson.Content = comment.Content
		commentJson.CreateDate = comment.CreatedAt.Format("2023-02-01 05:21")

		commentJsonList = append(commentJsonList, commentJson)
	}
	c.JSON(http.StatusOK, CommentListResponse{
		Response:    Response{StatusCode: 0},
		CommentList: commentJsonList,
	})
}
