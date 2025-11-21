package service

import (
	"checkingsocial/farcaster"
	"checkingsocial/internal/model"
	"checkingsocial/twitter"
	"errors"
	"fmt"
	"sync"
)

// Checker định nghĩa interface cho việc kiểm tra tài khoản mạng xã hội.
type Checker interface {
	Check(req model.CheckRequest) model.CheckResponse
	BatchCheck(req model.BatchCheckRequest) model.BatchCheckResponse
	CheckSocialAction(req model.SocialActionRequest) (bool, error)
}

// socialChecker là implementation của Checker.
type socialChecker struct{}

// NewSocialChecker tạo một instance mới của socialChecker.
func NewSocialChecker() Checker {
	return &socialChecker{}
}

// CheckSocialAction thực hiện kiểm tra một hành động xã hội.
func (s *socialChecker) CheckSocialAction(req model.SocialActionRequest) (bool, error) {
	if req.Social == "farcaster" && req.Action == "follow" {
		return farcaster.CheckFollow(req.IDUser)
	}
	if req.Social == "x" && req.Action == "follow" {
		return twitter.CheckFollow(req.IDUser)
	}
	return false, errors.New("unsupported social or action")
}

// Check thực hiện kiểm tra một tài khoản mạng xã hội.
// TODO: Implement a real check logic for each platform.
func (s *socialChecker) Check(req model.CheckRequest) model.CheckResponse {
	// Giả lập logic: hiện tại luôn trả về tồn tại
	exists := true
	profileURL := fmt.Sprintf("https://www.%s.com/%s", req.Platform, req.Username)

	return model.CheckResponse{
		Platform:   req.Platform,
		Username:   req.Username,
		Exists:     exists,
		ProfileURL: profileURL,
		Message:    "Checked successfully (simulated).",
	}
}

// BatchCheck thực hiện kiểm tra nhiều tài khoản cùng lúc sử dụng goroutines.
func (s *socialChecker) BatchCheck(req model.BatchCheckRequest) model.BatchCheckResponse {
	var wg sync.WaitGroup
	resultsChan := make(chan model.CheckResponse, len(req.Checks))

	for _, checkReq := range req.Checks {
		wg.Add(1)
		go func(cr model.CheckRequest) {
			defer wg.Done()
			resultsChan <- s.Check(cr)
		}(checkReq)
	}

	wg.Wait()
	close(resultsChan)

	var results []model.CheckResponse
	successCount := 0
	for res := range resultsChan {
		results = append(results, res)
		if res.Exists {
			successCount++
		}
	}

	return model.BatchCheckResponse{
		Results: results,
		Total:   len(req.Checks),
		Success: successCount,
		Failed:  len(req.Checks) - successCount,
	}
}
