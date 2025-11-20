package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// extractCookieValue returns the value for a cookie name from a raw Cookie header string.
// It trims spaces and surrounding quotes.
func extractCookieValue(cookieHeader, name string) string {
	parts := strings.Split(cookieHeader, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == name {
			return strings.Trim(v, "\"")
		}
	}
	return ""
}

// CheckFollow checks if the given userId follows the target Twitter account (by rest_id).
// It calls X (Twitter) internal GraphQL Following endpoint and scans the returned entries.
// Returns true if an entryId of form "user-<TARGET_ID>" is present in the returned page; otherwise false.
func CheckFollow(userID string) (bool, error) {
	// Load env vars if .env present
	_ = godotenv.Load()

	// Required configuration
	queryID := os.Getenv("TWITTER_QUERY_ID")
	if queryID == "" {
		// default to a commonly observed hash, but allow override by env as this changes often
		queryID = "S5xUN9s2v4xk50KWGGvyvQ"
	}
	bearer := os.Getenv("TWITTER_BEARER")
	cookie := os.Getenv("TWITTER_COOKIE")
	csrf := os.Getenv("TWITTER_CT0")
	targetID := os.Getenv("TWITTER_TARGET_ID")

	if bearer == "" {
		return false, fmt.Errorf("missing required envs: TWITTER_BEARER")
	}
	// Decode cookie if user pasted URL-encoded value (e.g., with %3A, %3B, etc.)
	if dec, err := url.QueryUnescape(cookie); err == nil {
		cookie = dec
	}
	// If CT0 not provided separately, try to extract from cookie string
	if csrf == "" {
		csrf = extractCookieValue(cookie, "ct0")
	}
	if csrf == "" {
		return false, fmt.Errorf("missing CSRF token: set TWITTER_CT0 or include ct0 in TWITTER_COOKIE")
	}
	if targetID == "" {
		return false, fmt.Errorf("TWITTER_TARGET_ID not set")
	}

	variables := map[string]any{
		"userId":                 userID,
		"count":                  20,
		"includePromotedContent": false,
		"withGrokTranslatedBio":  false,
	}
	features := map[string]any{
		"rweb_video_screen_enabled":                                               false,
		"profile_label_improvements_pcf_label_in_post_enabled":                    true,
		"responsive_web_profile_redirect_enabled":                                 false,
		"rweb_tipjar_consumption_enabled":                                         true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"premium_content_api_read_enabled":                                        false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"responsive_web_grok_analyze_button_fetch_trends_enabled":                 false,
		"responsive_web_grok_analyze_post_followups_enabled":                      true,
		"responsive_web_jetfuel_frame":                                            true,
		"responsive_web_grok_share_attachment_enabled":                            true,
		"articles_preview_enabled":                                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"responsive_web_grok_show_grok_translated_post":                           false,
		"responsive_web_grok_analysis_button_from_backend":                        true,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_grok_image_annotation_enabled":                            true,
		"responsive_web_grok_imagine_annotation_enabled":                          true,
		"responsive_web_grok_community_note_auto_translation_is_enabled":          false,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	variablesJSON, _ := json.Marshal(variables)
	featuresJSON, _ := json.Marshal(features)

	u := fmt.Sprintf("https://x.com/i/api/graphql/%s/Following?variables=%s&features=%s",
		queryID, url.QueryEscape(string(variablesJSON)), url.QueryEscape(string(featuresJSON)))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return false, err
	}

	// Headers (keep minimal required set)
	req.Header.Set("accept", "*/*")
	req.Header.Set("authorization", "Bearer "+bearer)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-csrf-token", csrf)
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	req.Header.Set("cookie", `guest_id_marketing=v1%3A176362541654113270; guest_id_ads=v1%3A176362541654113270; guest_id=v1%3A176362541654113270; personalization_id="v1_+w7xL28oM+zyv6l07Acgcg=="; gt=1991415449637253140; __cuid=f870393e1dcd4ce5ac31294aa53e0d25; g_state={"i_l":0,"i_ll":1763625531382,"i_b":"fbktveH4t9AiDD8L295e3ciN7z/D7zvUPKRqywJV3Cg"}; kdt=pMx5PX08CWcMA9dDv8Wfjy7hCpGSdmZ1dQStT6IQ; auth_token=cafa295c85e1fcb6cade0e080be8547b153eb0fb; ct0=abfe3ce07dd4679919844e313099cf2ced7e97164f1042782a72c76db1206c89841ff0dfa4b2f766d82affb14c3fe985ae516249deffe92d184752915b5ce2b8b540d03b6b944051f7f059c714186426; att=1-QmzcsPkP2JiwQLolyemVB2XB15NFhCRX336WlSAs; lang=en; twid=u%3D1750722368132308992; first_ref=https%3A%2F%2Fx.com%2Fi%2Fflow%2Flogin; __cf_bm=G0_ck9tVVhaAlDV.x_new0RmeEdiEapRcAYZi3mvfZ0-1763625890.730432-1.0.1.1-kz5ZJFPSX8iFEDhq_53GnwqTk3CvrJG7irHsg_rAdRXUSpx4d1sVIr4JuIfPuzYzyo9SW9.YtqTmxt.5EAt6PcdLyZhL4QhYgLEOhcUcBUU94FOxphsWX_nWYQkUypAq`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("twitter api status %d: %s", resp.StatusCode, string(body))
	}

	// Parse just enough of the response to find entryIds
	var parsed struct {
		Data struct {
			User struct {
				Result struct {
					Timeline struct {
						Timeline struct {
							Instructions []struct {
								Type    string `json:"type"`
								Entries []struct {
									EntryId string `json:"entryId"`
								} `json:"entries"`
							} `json:"instructions"`
						} `json:"timeline"`
					} `json:"timeline"`
				} `json:"result"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		// If unmarshal fails, return detailed snippet for debugging
		snippet := string(body)
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		return false, fmt.Errorf("failed to parse twitter response: %v; body: %s", err, snippet)
	}

	prefix := "user-" + targetID
	for _, instr := range parsed.Data.User.Result.Timeline.Timeline.Instructions {
		if strings.EqualFold(instr.Type, "TimelineAddEntries") {
			for _, e := range instr.Entries {
				if e.EntryId == prefix {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
