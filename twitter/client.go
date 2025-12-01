package twitter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// isTruthy returns true if an env-like string represents a truthy value
func isTruthy(s string) bool {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	}
	return false
}

// CheckFollow checks if user_b (userID) follows user_a (target username) using Apify actor.
// Inputs:
//   - userID: will be sent as user_b (the account to check if it follows target)
//
// Config via ENV:
//   - TWITTER_TARGET_USERNAME: the target account (user_a)
//   - APIFY_ACT_URL (optional): override Apify actor URL (defaults to UC0t7r32caYf7tYgZ)
//   - X_TOKEN_FILE (optional): path to x.txt (default: ./x.txt) with lines: cookie|token
//
// Behavior:
//   - Picks a random cookie|token pair from x.txt
//   - Calls Apify run-sync-get-dataset-items with JSON body
//   - Returns true if user_b_follows_user_a is true in the first item of result
func CheckFollow(userID string) (bool, error) {
	_ = godotenv.Load()

	target := os.Getenv("TWITTER_TARGET_USERNAME")
	if target == "" {
		return false, errors.New("TWITTER_TARGET_USERNAME not set")
	}

	apifyURL := os.Getenv("APIFY_ACT_URL")
	if apifyURL == "" {
		apifyURL = "https://api.apify.com/v2/acts/UC0t7r32caYf7tYgZ/run-sync-get-dataset-items"
	}

	cookies := "[{\"domain\":\".x.com\",\"expirationDate\":1746856132.619477,\"hostOnly\":false,\"httpOnly\":true,\"name\":\"__cf_bm\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"nxqSkSFt_3UkBWH576nmL44QmPDWzVDjgIkLfeLHOsM-1746854332-1.0.1.1-.D.PjFVYMhjL77et4PsMZlLNYXg4KRzBUFRqAWpRcQbXf6H_5DIZQulrUK7tu34y3rdkEuyq6rjjJfYYxoXEs5vG9W8r.OWN.0WehXsCBOU\",\"id\":1},{\"domain\":\".x.com\",\"expirationDate\":1781195046.756188,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"_ga\",\"path\":\"/\",\"sameSite\":\"unspecified\",\"secure\":false,\"session\":false,\"storeId\":\"0\",\"value\":\"GA1.1.798476497.1746635047\",\"id\":2},{\"domain\":\".x.com\",\"expirationDate\":1781195575.696591,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"_ga_RJGMY4G45L\",\"path\":\"/\",\"sameSite\":\"unspecified\",\"secure\":false,\"session\":false,\"storeId\":\"0\",\"value\":\"GS2.1.s1746635046$o1$g0$t1746635575$j60$l0$h0\",\"id\":3},{\"domain\":\".x.com\",\"expirationDate\":1771085083.01971,\"hostOnly\":false,\"httpOnly\":true,\"name\":\"auth_token\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"e8ebde28ed35f60bbaa480fadf171a2706262770\",\"id\":4},{\"domain\":\".x.com\",\"expirationDate\":1771085083.370993,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"ct0\",\"path\":\"/\",\"sameSite\":\"lax\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"b49f42546ed0263e7ed05dd736255e27cb99c44fb82abe573b42ae3132c7a5272856579502280f5624892335fcc0c6a686836b689565d9d0aaf4ad2c630fb7d77a67b7863020e130eeb1ed7dbf907a0c\",\"id\":5},{\"domain\":\".x.com\",\"expirationDate\":1767339612.610054,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"guest_id\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"v1%3A172490945488285787\",\"id\":6},{\"domain\":\".x.com\",\"expirationDate\":1781414587.680152,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"guest_id_ads\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"v1%3A172490945488285787\",\"id\":7},{\"domain\":\".x.com\",\"expirationDate\":1781414587.680395,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"guest_id_marketing\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"v1%3A172490945488285787\",\"id\":8},{\"domain\":\".x.com\",\"expirationDate\":1771085083.019629,\"hostOnly\":false,\"httpOnly\":true,\"name\":\"kdt\",\"path\":\"/\",\"sameSite\":\"unspecified\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"0MprSsNU5W2zirK0Kdo6YW1QFh58oyeb4Qw5RChk\",\"id\":9},{\"domain\":\".x.com\",\"expirationDate\":1778390588.734748,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"night_mode\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"2\",\"id\":10},{\"domain\":\".x.com\",\"expirationDate\":1768061146.52658,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"personalization_id\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"\\\"v1_fo7TidorU55HXywHnSY3Yw==\\\"\",\"id\":11},{\"domain\":\".x.com\",\"expirationDate\":1778390593.659738,\"hostOnly\":false,\"httpOnly\":false,\"name\":\"twid\",\"path\":\"/\",\"sameSite\":\"no_restriction\",\"secure\":true,\"session\":false,\"storeId\":\"0\",\"value\":\"u%3D1877748387220799489\",\"id\":12},{\"domain\":\"x.com\",\"hostOnly\":true,\"httpOnly\":false,\"name\":\"lang\",\"path\":\"/\",\"sameSite\":\"unspecified\",\"secure\":false,\"session\":true,\"storeId\":\"0\",\"value\":\"en\",\"id\":13}]"
	token := "REMOVED_APIFY_TOKEN"

	// Build top-level payload as actor expects (no "input" wrapper)
	// Ensure cookies string is not double-quoted; unquote if needed
	if strings.HasPrefix(cookies, "\"") && strings.HasSuffix(cookies, "\"") {
		if unq, err := strconv.Unquote(cookies); err == nil {
			cookies = unq
		}
	}
	payload := struct {
		Cookies string `json:"cookies"`
		UserA   string `json:"user_a"`
		UserB   string `json:"user_b"`
	}{
		Cookies: cookies,
		UserA:   target,
		UserB:   userID,
	}
	bodyBytes, _ := json.Marshal(payload)

	// Debug payload (masked) if APIFY_DEBUG is enabled
	if isTruthy(os.Getenv("APIFY_DEBUG")) {
		masked := map[string]any{
			"cookies": "<masked>",
			"user_a":  target,
			"user_b":  userID,
		}
		mb, _ := json.Marshal(masked)
		log.Printf("[Apify] URL=%s", apifyURL)
		log.Printf("[Apify] Request=%s", string(mb))
	}

	req, err := http.NewRequest("POST", apifyURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read full body for flexible decoding and better error messages
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("apify read body error: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Log masked request for troubleshooting (top-level payload)
		masked := map[string]any{
			"cookies": "<masked>",
			"user_a":  target,
			"user_b":  userID,
		}
		mb, _ := json.Marshal(masked)
		log.Printf("[Apify][ERROR] URL=%s", apifyURL)
		log.Printf("[Apify][ERROR] Request=%s", string(mb))

		// Return a short snippet to avoid logging secrets
		snippet := string(body)
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		log.Printf("[Apify][ERROR] Status=%d Body=%s", resp.StatusCode, snippet)
		return false, fmt.Errorf("apify status %d: %s", resp.StatusCode, snippet)
	}

	// Try decode as array first (expected shape)
	type item struct {
		Status            string `json:"status"`
		UserBFollowsUserA bool   `json:"user_b_follows_user_a"`
	}
	var arr []item
	if err := json.Unmarshal(body, &arr); err == nil {
		if len(arr) == 0 {
			return false, errors.New("apify empty result")
		}
		return arr[0].UserBFollowsUserA, nil
	}

	// Fallback: sometimes actor returns a single object instead of array
	var obj item
	if err := json.Unmarshal(body, &obj); err == nil && (obj.Status != "" || obj.UserBFollowsUserA || strings.Contains(string(body), "user_b_follows_user_a")) {
		return obj.UserBFollowsUserA, nil
	}

	// Fallback 2: generic map to probe key
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil {
		if v, ok := m["user_b_follows_user_a"]; ok {
			if b, ok2 := v.(bool); ok2 {
				return b, nil
			}
		}
		if msg, ok := m["error"]; ok {
			return false, fmt.Errorf("apify error: %v", msg)
		}
	}

	snippet := string(body)
	if len(snippet) > 512 {
		snippet = snippet[:512]
	}
	return false, fmt.Errorf("apify decode error: body: %s", snippet)
}

// pickRandomCookieToken returns [cookies, token] by reading a random line from x.txt (or X_TOKEN_FILE)
// Each line format: cookie|token
func pickRandomCookieToken() ([2]string, error) {
	filePath := os.Getenv("X_TOKEN_FILE")
	if filePath == "" {
		filePath = "x.txt"
	}
	abs, _ := filepath.Abs(filePath)
	f, err := os.Open(abs)
	if err != nil {
		return [2]string{}, fmt.Errorf("open %s: %w", abs, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var pairs [][2]string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		pairs = append(pairs, [2]string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
	}
	if err := scanner.Err(); err != nil {
		return [2]string{}, err
	}
	if len(pairs) == 0 {
		return [2]string{}, errors.New("no cookie|token pair found in x.txt")
	}

	rand.Seed(time.Now().UnixNano())
	pick := pairs[rand.Intn(len(pairs))]
	return pick, nil
}
