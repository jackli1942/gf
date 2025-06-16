package i18nutil

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv" // For gconv.String()
)

// T translates a content by given key and optional arguments.
func T(ctx context.Context, key string, GfGenericVariables ...interface{}) string {
    rawString := gi18n.Instance().T(ctx, key)

    if len(GfGenericVariables) > 0 {
        // Attempt to use as {key} style placeholders with a map
        if gstr.Contains(rawString, "{") && gstr.Contains(rawString, "}") {
            varsMap := make(g.Map)
            processedMap := false

            if len(GfGenericVariables) == 1 {
                if m, ok := GfGenericVariables[0].(g.Map); ok {
                    varsMap = m
                    processedMap = true
                } else if mm, ok := GfGenericVariables[0].(map[string]interface{}); ok {
                    varsMap = mm
                    processedMap = true
                }
            } else if len(GfGenericVariables)%2 == 0 { // Key-value pairs
                tempMap := make(g.Map, len(GfGenericVariables)/2)
                validPairs := true
                for i := 0; i < len(GfGenericVariables); i += 2 {
                    if keyStr, ok := GfGenericVariables[i].(string); ok {
                        tempMap[keyStr] = GfGenericVariables[i+1]
                    } else {
                        g.Log().Warningf(ctx, "i18nutil.T: Placeholder key at index %d for key '%s' is not a string. Map replacement might be incomplete.", i, key)
                        validPairs = false
                        break
                    }
                }
                if validPairs {
                    varsMap = tempMap
                    processedMap = true
                }
            }

            if processedMap && len(varsMap) > 0 {
                tempStr := rawString
                for k, v := range varsMap {
                    placeholder := "{" + k + "}"
                    tempStr = gstr.Replace(tempStr, placeholder, gconv.String(v))
                }
                return tempStr
            }
        }

        // Fallback for fmt style placeholders (%s, %d) if no map replacement occurred
        if gstr.Contains(rawString, "%") {
             g.Log().Debugf(ctx, "i18nutil.T: Attempting fmt.Sprintf for key '%s' as {key} replacement did not occur or was not applicable.", key)
             return fmt.Sprintf(rawString, GfGenericVariables...)
        }
    }
    return rawString
}
