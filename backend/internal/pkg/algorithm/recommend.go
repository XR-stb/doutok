package algorithm

import (
	"math"
	"sort"
	"time"
)

// ==================== 视频推荐算法 ====================

// VideoFeature 视频特征向量
type VideoFeature struct {
	VideoID      int64
	AuthorID     int64
	Tags         []string
	LikeRate     float64 // 点赞率
	CommentRate  float64 // 评论率
	ShareRate    float64 // 分享率
	CompletionRate float64 // 完播率
	PublishTime  time.Time
	Score        float64 // 最终推荐分
}

// UserProfile 用户兴趣画像
type UserProfile struct {
	UserID        int64
	TagWeights    map[string]float64 // 标签兴趣权重
	AuthorWeights map[int64]float64  // 作者偏好权重
	AvgWatchTime  float64            // 平均观看时长
	ActiveHours   []int              // 活跃时段
}

// FeedRecommender 推荐Feed流引擎
type FeedRecommender struct {
	// 权重配置
	interestWeight    float64
	hotWeight         float64
	freshWeight       float64
	diversityWeight   float64
	socialWeight      float64
}

func NewFeedRecommender() *FeedRecommender {
	return &FeedRecommender{
		interestWeight:  0.4,
		hotWeight:       0.25,
		freshWeight:     0.15,
		diversityWeight: 0.1,
		socialWeight:    0.1,
	}
}

// Rank 多因子打分排序
func (r *FeedRecommender) Rank(user *UserProfile, candidates []VideoFeature, followingIDs []int64) []VideoFeature {
	followSet := make(map[int64]bool, len(followingIDs))
	for _, id := range followingIDs {
		followSet[id] = true
	}

	for i := range candidates {
		v := &candidates[i]

		// 1. 兴趣匹配分 - 基于用户标签画像的余弦相似度
		interestScore := r.calcInterestScore(user, v)

		// 2. 热度分 - 基于互动率的威尔逊区间下界
		hotScore := r.calcHotScore(v)

		// 3. 新鲜度分 - 时间衰减
		freshScore := r.calcFreshScore(v.PublishTime)

		// 4. 社交分 - 关注的人发的内容加权
		socialScore := 0.0
		if followSet[v.AuthorID] {
			socialScore = 1.0
		}
		if w, ok := user.AuthorWeights[v.AuthorID]; ok {
			socialScore = math.Max(socialScore, w)
		}

		v.Score = r.interestWeight*interestScore +
			r.hotWeight*hotScore +
			r.freshWeight*freshScore +
			r.socialWeight*socialScore
	}

	// 5. 多样性打散 - 相同作者/标签不连续出现
	result := r.diversify(candidates)
	return result
}

func (r *FeedRecommender) calcInterestScore(user *UserProfile, v *VideoFeature) float64 {
	if len(v.Tags) == 0 || len(user.TagWeights) == 0 {
		return 0.5
	}
	var sum float64
	var count int
	for _, tag := range v.Tags {
		if w, ok := user.TagWeights[tag]; ok {
			sum += w
			count++
		}
	}
	if count == 0 {
		return 0.3 // 探索性推荐
	}
	return math.Min(sum/float64(count), 1.0)
}

func (r *FeedRecommender) calcHotScore(v *VideoFeature) float64 {
	// 威尔逊区间下界 - 解决小样本高点赞率问题
	n := v.LikeRate + v.CommentRate + v.ShareRate
	if n == 0 {
		return 0
	}
	p := v.LikeRate
	z := 1.96 // 95% confidence
	return (p + z*z/(2*n) - z*math.Sqrt((p*(1-p)+z*z/(4*n))/n)) / (1 + z*z/n)
}

func (r *FeedRecommender) calcFreshScore(publishTime time.Time) float64 {
	hours := time.Since(publishTime).Hours()
	// 指数衰减：半衰期 24 小时
	return math.Exp(-0.693 * hours / 24.0)
}

func (r *FeedRecommender) diversify(videos []VideoFeature) []VideoFeature {
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Score > videos[j].Score
	})

	if len(videos) <= 3 {
		return videos
	}

	result := make([]VideoFeature, 0, len(videos))
	lastAuthor := int64(0)
	skipQueue := make([]VideoFeature, 0)

	for _, v := range videos {
		if v.AuthorID == lastAuthor && len(result) > 0 {
			skipQueue = append(skipQueue, v)
		} else {
			result = append(result, v)
			lastAuthor = v.AuthorID
			// 插入一个之前跳过的
			if len(skipQueue) > 0 {
				result = append(result, skipQueue[0])
				skipQueue = skipQueue[1:]
			}
		}
	}
	result = append(result, skipQueue...)
	return result
}

// ==================== 评论排序算法 ====================

// CommentFeature 评论特征
type CommentFeature struct {
	CommentID  int64
	UserID     int64
	LikeCount  int64
	ReplyCount int64
	CreatedAt  time.Time
	IsAuthor   bool // 是否是视频作者
	Score      float64
}

// CommentRanker 评论排序器
type CommentRanker struct{}

func NewCommentRanker() *CommentRanker {
	return &CommentRanker{}
}

// RankComments 评论排序 - 热度 + 时间 + 身份加权
func (cr *CommentRanker) RankComments(comments []CommentFeature, videoAuthorID int64) []CommentFeature {
	for i := range comments {
		c := &comments[i]

		// 互动热度 (对数平滑)
		hotScore := math.Log2(float64(c.LikeCount+1)) + 0.5*math.Log2(float64(c.ReplyCount+1))

		// 时间衰减 (半衰期 12 小时)
		hours := time.Since(c.CreatedAt).Hours()
		timeScore := math.Exp(-0.693 * hours / 12.0)

		// 身份加权
		authorBoost := 0.0
		if c.UserID == videoAuthorID {
			authorBoost = 3.0 // 视频作者评论置顶加权
		}

		c.Score = hotScore*0.6 + timeScore*0.3 + authorBoost*0.1
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].Score > comments[j].Score
	})
	return comments
}

// ==================== 直播排行榜 ====================

// LiveRankEntry 直播排行条目
type LiveRankEntry struct {
	UserID   int64   `json:"user_id"`
	Nickname string  `json:"nickname"`
	Avatar   string  `json:"avatar"`
	Amount   int64   `json:"amount"` // 送礼总额
	Score    float64 `json:"score"`
}

// LiveRanker 直播排行计算器
type LiveRanker struct{}

func NewLiveRanker() *LiveRanker {
	return &LiveRanker{}
}

// CalcRank 计算直播排行榜
// 使用 Redis ZSET 实现实时排行，这里是分数计算逻辑
func (lr *LiveRanker) CalcScore(giftAmount int64, comboCount int, isFirstGift bool) float64 {
	base := float64(giftAmount)

	// 连击加成：combo越多分数越高
	comboMultiplier := 1.0 + math.Log2(float64(comboCount+1))*0.1

	// 首次送礼加成
	firstBonus := 0.0
	if isFirstGift {
		firstBonus = float64(giftAmount) * 0.2
	}

	return base*comboMultiplier + firstBonus
}
