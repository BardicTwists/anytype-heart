package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/anyproto/anytype-heart/pkg/lib/logging"
)

const (
	baseURL = "http://localhost:31009/v1"
	// testSpaceId = "bafyreifymx5ucm3fdc7vupfg7wakdo5qelni3jvlmawlnvjcppurn2b3di.2lcu0r85yg10d" // dev (entry space)
	// testSpaceId = "bafyreiezhzb4ggnhjwejmh67pd5grilk6jn3jt7y2rnfpbkjwekilreola.1t123w9f2lgn5" // LFLC
	// testSpaceId  = "bafyreiakofsfkgb7psju346cir2hit5hinhywaybi6vhp7hx4jw7hkngje.scoxzd7vu6rz" // HPI
	// testObjectId = "bafyreidhtlbbspxecab6xf4pi5zyxcmvwy6lqzursbjouq5fxovh6y3xwu"              // "Work Faster with Templates"
	// testObjectId = "bafyreib3i5uq2tztocw3wrvhdugkwoxgg2xjh2jnl5retnyky66mr5b274" // Tag Test Page (in dev space)
	// testTypeId   = "bafyreie3djy4mcldt3hgeet6bnjay2iajdyi2fvx556n6wcxii7brlni3i" // Page (in dev space)
	// chatSpaceId  = "bafyreigryvrmerbtfswwz5kav2uq5dlvx3hl45kxn4nflg7lz46lneqs7m.2nvj2qik6ctdy" // Anytype Wiki space
	// chatSpaceId = "bafyreiexhpzaf7uxzheubh7cjeusqukjnxfvvhh4at6bygljwvto2dttnm.2lcu0r85yg10d" // chat space
)

var log = logging.Logger("rest-api")

// ReplacePlaceholders replaces placeholders in the endpoint with actual values from parameters.
func ReplacePlaceholders(endpoint string, parameters map[string]interface{}) string {
	for key, value := range parameters {
		placeholder := fmt.Sprintf("{%s}", key)
		encodedValue := url.QueryEscape(fmt.Sprintf("%v", value))
		endpoint = strings.ReplaceAll(endpoint, placeholder, encodedValue)
	}

	// Parse the base URL + endpoint
	u, err := url.Parse(baseURL + endpoint)
	if err != nil {
		log.Errorf("Failed to parse URL: %v\n", err)
		return ""
	}

	return u.String()
}

func main() {
	endpoints := []struct {
		method     string
		endpoint   string
		parameters map[string]interface{}
		body       map[string]interface{}
	}{
		// auth
		// {"POST", "/auth/displayCode", nil, nil},
		// {"GET", "/auth/token?challengeId={challengeId}&code={code}", map[string]interface{}{"challengeId": "6738dfc5cda913aad90e8c2a", "code": "2931"}, nil},

		// export
		// {"GET", "/spaces/{space_id}/objects/{object_id}/export/{format}", map[string]interface{}{"space_id": testSpaceId, "object_id": testObjectId, "format": "markdown"}, nil},

		// spaces
		// {"POST", "/spaces", nil, map[string]interface{}{"name": "New Space"}},
		// {"GET", "/spaces?limit={limit}&offset={offset}", map[string]interface{}{"limit": 100, "offset": 0}, nil},
		// {"GET", "/spaces/{space_id}/members?limit={limit}&offset={offset}", map[string]interface{}{"space_id": testSpaceId, "limit": 100, "offset": 0}, nil},

		// objects
		// {"GET", "/spaces/{space_id}/objects?limit={limit}&offset={offset}", map[string]interface{}{"space_id": testSpaceId, "limit": 100, "offset": 0}, nil},
		// {"GET", "/spaces/{space_id}/objects/{object_id}", map[string]interface{}{"space_id": testSpaceId, "object_id": testObjectId}, nil},
		// {"POST", "/spaces/{space_id}/objects", map[string]interface{}{"space_id": testSpaceId}, map[string]interface{}{"name": "New Object from demo", "icon": "💥", "template_id": "", "object_type_unique_key": "ot-page", "with_chat": false}},
		// {"PUT", "/spaces/{space_id}/objects/{object_id}", map[string]interface{}{"space_id": testSpaceId, "object_id": testObjectId}, map[string]interface{}{"name": "Updated Object"}},
		// {"GET", "/spaces/{space_id}/object_types?limit={limit}&offset={offset}", map[string]interface{}{"space_id": testSpaceId, "limit": 100, "offset": 0}, nil},
		// {"GET", "/spaces/{space_id}/object_types/{type_id}/templates?limit={limit}&offset={offset}", map[string]interface{}{"space_id": testSpaceId, "type_id": testTypeId, "limit": 100, "offset": 0}, nil},

		// search
		// {"GET", "/search?query={query}&object_types={object_types}&limit={limit}&offset={offset}", map[string]interface{}{"query": "new", "object_types": "", "limit": 100, "offset": 0}, nil},
	}

	for _, ep := range endpoints {
		finalURL := ReplacePlaceholders(ep.endpoint, ep.parameters)

		var req *http.Request
		var err error

		if ep.body != nil {
			body, err := json.Marshal(ep.body)
			if err != nil {
				log.Errorf("Failed to marshal body for %s: %v\n", ep.endpoint, err)
				continue
			}
			req, err = http.NewRequest(ep.method, finalURL, bytes.NewBuffer(body))
			if err != nil {
				log.Errorf("Failed to create request for %s: %v\n", ep.endpoint, err)
			}
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequest(ep.method, finalURL, nil)
		}

		if err != nil {
			log.Errorf("Failed to create request for %s: %v\n", ep.endpoint, err)
			continue
		}

		// Execute the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Errorf("Failed to make request to %s: %v\n", finalURL, err.Error())
			continue
		}
		defer resp.Body.Close()

		// Check the status code
		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("Failes to read response body for request to %s with code %d.", finalURL, resp.StatusCode)
				continue
			}
			log.Errorf("Request to %s returned status code %d: %v\n", finalURL, resp.StatusCode, string(body))
			continue
		}

		// Read the response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Failed to read response body for %s: %v\n", ep.endpoint, err)
			continue
		}

		// Log the response
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, body, "", "  ")
		if err != nil {
			log.Errorf("Failed to pretty print response body for %s: %v\n", ep.endpoint, err)
			log.Infof("%s\n", string(body))
			continue
		}

		log.Infof("Endpoint: %s, Status Code: %d, Body: %s\n", finalURL, resp.StatusCode, prettyJSON.String())
	}
}
