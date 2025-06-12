package health_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
	"yuncms/internal/router"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	"github.com/gogf/gf/v2/frame/g"
	// "github.com/gogf/gf/v2/i18n/gi18n" // Avoid if GetLanguage/GetPath are undefined
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/stretchr/testify/assert"
)

func getEnvOrFail(t *testing.T, key string) string {
    val := os.Getenv(key)
    if val == "" {
        t.Fatalf("Environment variable %s is not set. This test relies on the subtask runner to set it.", key)
    }
    return val
}

func TestHealthEndpoint(t *testing.T) {
    tempConfigPath := getEnvOrFail(t, "YUNCMS_TEST_CONFIG_PATH")
    tempI18nDir := getEnvOrFail(t, "YUNCMS_TEST_I18N_PATH")

	originalAdapter := g.Cfg().GetAdapter()
	testAdapter, err := gcfg.NewAdapterFile(tempConfigPath)
    if err != nil {
        t.Fatalf("Failed to create adapter from temp config file '%s': %v", tempConfigPath, err)
    }
	g.Cfg().SetAdapter(testAdapter)
	defer g.Cfg().SetAdapter(originalAdapter)

	// Workaround for GetPath/GetLanguage issues:
	// originalI18nPath := g.I18n().GetPath()
	// originalI18nLang := g.I18n().GetLanguage()

	if err := g.I18n().SetPath(tempI18nDir); err != nil {
	    t.Fatalf("Test setup: Failed to set i18n path to '%s': %v", tempI18nDir, err)
	}
	g.I18n().SetLanguage("zh") // Default is now Chinese as per previous subtask
	defer func() {
	    g.I18n().SetPath("")       // Reset path to empty
	    g.I18n().SetLanguage("zh") // Restore to known default "zh" (aligns with test logic)
	}()

	s := g.Server("testHealthEndpointServerFullChecks")
	router.BindController(s)
	s.SetAddr("127.0.0.1:0")

    go func() {
        if err := s.Start(); err != nil {
            g.Log().Errorf(context.Background(), "TestHealthEndpoint server failed to start: %v", err)
        }
    }()
	defer s.Shutdown()

    var realPort int
    for i := 0; i < 20; i++ {
        realPort = s.GetListenedPort()
        if realPort > 0 {
            break
        }
        time.Sleep(100 * time.Millisecond)
    }
    if realPort <= 0 {
        t.Fatalf("TestHealthEndpoint server did not start listening on a valid port. Port: %d", realPort)
    }
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", realPort)

	testCases := []struct {
		name                 string
		langQuery            string
		expectedStatus       string
		expectedLangUsed     string
		expectedMsgSubstring string
		expectDBError        bool
		expectRedisError     bool
	}{
		{"default lang (zh)", "", "ok_WriteJsonExit_full_checks", "zh", "Yuncms 服务正在运行 (zh)", true, true},
		{"english lang", "?lang=en", "ok_WriteJsonExit_full_checks", "en", "Yuncms service is running (en)", true, true},
		{"chinese lang", "?lang=zh", "ok_WriteJsonExit_full_checks", "zh", "Yuncms 服务正在运行 (zh)", true, true},
		{"unsupported lang", "?lang=fr", "ok_WriteJsonExit_full_checks", "zh", "Yuncms 服务正在运行 (zh)", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqURL := baseURL + "/health" + tc.langQuery
			resp, err := http.Get(reqURL)
			assert.NoError(t, err, "HTTP GET request should not fail")
			if err != nil { return }
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK")

			bodyBytes, readErr := ioutil.ReadAll(resp.Body)
			assert.NoError(t, readErr, "Should be able to read body")

			assert.NotEmpty(t, bodyBytes, "Response body SHOULD NOT BE EMPTY. URL: %s", reqURL)
			if len(bodyBytes) == 0 { return }

			var healthData g.Map
			err = json.Unmarshal(bodyBytes, &healthData)
			assert.NoError(t, err, "JSON decoding should not fail. Body: %s", string(bodyBytes))

			assert.Equal(t, tc.expectedStatus, healthData["status"], "Status field mismatch. Body: %s", string(bodyBytes))
			assert.Equal(t, tc.expectedLangUsed, healthData["lang_used"], "lang_used field mismatch. Body: %s", string(bodyBytes))
			assert.Contains(t, healthData["message"], tc.expectedMsgSubstring, "Message content mismatch. Body: %s", string(bodyBytes))

            expectedLangRequested := ""
            if tc.langQuery != "" && strings.HasPrefix(tc.langQuery, "?lang=") {
                expectedLangRequested = strings.TrimPrefix(tc.langQuery, "?lang=")
            }
            assert.Equal(t, expectedLangRequested, healthData["lang_requested"], "lang_requested mismatch. Body: %s", string(bodyBytes))

			dbStatus, _ := healthData["dbStatus"].(string)
			redisStatus, _ := healthData["redisStatus"].(string)

			if tc.expectDBError {
				assert.Contains(t, dbStatus, "error", "Expected DB status to indicate an error. Got: %s. Body: %s", dbStatus, string(bodyBytes))
			} else {
				assert.Equal(t, "ok", dbStatus, "Expected DB status to be 'ok'. Got: %s. Body: %s", dbStatus, string(bodyBytes))
			}
			if tc.expectRedisError {
				assert.Contains(t, redisStatus, "error", "Expected Redis status to indicate an error. Got: %s. Body: %s", redisStatus, string(bodyBytes))
			} else {
				assert.Equal(t, "ok", redisStatus, "Expected Redis status to be 'ok'. Got: %s. Body: %s", redisStatus, string(bodyBytes))
			}
		})
	}
}

func TestMinimalServerHealth(t *testing.T) {
	s := g.Server(grand.S(10))
	s.SetDumpRouterMap(false)
	logger := g.Log(grand.S(5))
	s.SetLogger(logger)

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.GET("/minimalhealth", func(r *ghttp.Request) {
			s.Logger().Info(r.Context(), "Minimal health handler reached in TestMinimalServerHealth!")
			r.Response.WriteJson(g.Map{"ping": "pong"})
		})
	})

	s.SetAddr("127.0.0.1:0")
    go func() {
        if err := s.Start(); err != nil {
            g.Log().Errorf(context.Background(), "MinimalServer server failed to start: %v", err)
        }
    }()
	defer s.Shutdown()

    var realPort int
    for i := 0; i < 20; i++ {
        realPort = s.GetListenedPort()
        if realPort > 0 {
            break
        }
        time.Sleep(100 * time.Millisecond)
    }
    if realPort <= 0 {
        t.Fatalf("MinimalServer server did not start listening on a valid port. Port: %d", realPort)
    }
	reqURL := fmt.Sprintf("http://127.0.0.1:%d/minimalhealth", realPort)

	s.Logger().Infof(context.Background(), "MinimalServer: Test GET %s", reqURL)

	resp, err := http.Get(reqURL)
	assert.NoError(t, err, "MinimalServer: HTTP GET request should not fail")
	if err != nil { return }
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "MinimalServer: HTTP status code should be 200 OK")

	bodyBytes, readErr := ioutil.ReadAll(resp.Body)
	assert.NoError(t, readErr, "MinimalServer: Should be able to read body")

	s.Logger().Infof(context.Background(), "MinimalServer: Response Body: %s", string(bodyBytes))
	assert.NotEmpty(t, bodyBytes, "MinimalServer: Response body SHOULD NOT BE EMPTY")

	if len(bodyBytes) > 0 {
		var result g.Map
		err = json.Unmarshal(bodyBytes, &result)
		assert.NoError(t, err, "MinimalServer: JSON decoding should not fail. Body: %s", string(bodyBytes))
		assert.Equal(t, "pong", result["ping"], "MinimalServer: 'ping' field should be 'pong'")
	} else {
        t.Logf("MinimalServer: Server Listeners: %v", s.GetListenedPorts())
        t.Logf("MinimalServer: Server Name: %s", s.GetName())
    }
}
