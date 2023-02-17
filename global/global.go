package global

import (
	"context"
	"github.com/Ljkkun/GreenBeanMiners/config"
	"github.com/go-redis/redis/v8"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	CONFIG               config.System            // 系统配置信息
	DB                   *gorm.DB                 // 数据库接口
	REDIS                *redis.Client            // Redis 缓存接口
	FILE_TYPE_MAP        sync.Map                 // 文件类型映射
	ID_GENERATOR         *sonyflake.Sonyflake     // 主键生成器
	CONTEXT              = context.Background()   // 上下文信息
	AUTO_CREATE_DB       = true                   // 是否自动生成数据库
	MAX_USERNAME_LENGTH  = 32                     // 用户名最大长度
	MIN_PASSWORD_PATTERN = "^[_a-zA-Z0-9]{6,32}$" // 密码格式
	START_TIME           = "2022-05-21 00:00:01"  // 固定启动时间，保证生成 ID 唯一性
	FEED_NUM             = 30                     // 每次返回视频数量
	VIDEO_ADDR           = "./public/video/"      // 视频存放位置
	COVER_ADDR           = "./public/cover/"      // 封面存放位置
	MAX_FILE_SIZE        = int64(10 << 20)        // 上传文件大小限制为10MB
	MAX_TITLE_LENGTH     = 140                    // 视频描述最大长度
	MAX_COMMENT_LENGTH   = 300                    // 评论最大长度
	WHITELIST_VIDEO      = map[string]bool{".mp4": true, ".avi": true, ".wmv": true, ".mpeg": true,
		".mov": true, ".flv": true, ".rmvb": true, ".3gb": true, ".vob": true, ".m4v": true}
)

// 过期时间
var (
	FAVORITE_EXPIRE       = 10 * time.Minute
	VIDEO_COMMENTS_EXPIRE = 10 * time.Minute
	COMMENT_EXPIRE        = 10 * time.Minute
	FOLLOW_EXPIRE         = 10 * time.Minute
	USER_INFO_EXPIRE      = 10 * time.Minute
	VIDEO_EXPIRE          = 10 * time.Minute
	PUBLISH_EXPIRE        = 10 * time.Minute
	EMPTY_EXPIRE          = 10 * time.Minute
	EXPIRE_TIME_JITTER    = 10 * time.Minute
)
