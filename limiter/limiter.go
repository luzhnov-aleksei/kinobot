package limiter

import (
	"time"
)

// лимит запросов в день от пользователя.
const maxMessagesPerDay = 20

type UserMessageInfo struct {
	Count   int
	ResetAt time.Time
}

var userMessages = make(map[int64]*UserMessageInfo)

func CanSendMessage(userID int64) bool {
	userInfo, exists := userMessages[userID]
	if !exists {
		userMessages[userID] = &UserMessageInfo{
			Count:   0,
			ResetAt: time.Now().Add(24 * time.Hour),
		}
		return true
	}

	if time.Now().After(userInfo.ResetAt) {
		userInfo.Count = 0
		userInfo.ResetAt = time.Now().Add(24 * time.Hour)
	}

	return userInfo.Count < maxMessagesPerDay
}

func IncrementMessageCount(userID int64) {
	userInfo := userMessages[userID]
	userInfo.Count++
}