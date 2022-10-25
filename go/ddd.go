//ddd

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	disconnectConversations("ssssss", *http.DefaultClient)
}

type ConversationsResponse struct {
	Conversations []struct {
		ConversationID string `json:"conversationId"`
	} `json:"conversations"`
}

func disconnectConversations(queueId string, apiClient *http.Client) {
	for true {
		query := "{\"paging\":{\"pageSize\":100,\"pageNumber\":1},\"order\":\"desc\",\"segmentFilters\":[{\"type\":\"and\",\"predicates\":[{\"type\":\"dimension\",\"dimension\":\"segmentType\",\"operator\":\"matches\",\"value\":\"interact\"},{\"type\":\"dimension\",\"dimension\":\"queueId\",\"operator\":\"matches\",\"value\":\"" + queueId + "\"},{\"type\":\"dimension\",\"dimension\":\"segmentEnd\",\"operator\":\"notExists\"}]}],\"conversationFilters\":[{\"type\":\"and\",\"predicates\":[{\"type\":\"dimension\",\"dimension\":\"conversationEnd\",\"operator\":\"notExists\"}]}],\"interval\":\"2018-01-07T19:42:31.141Z/2018-02-06T19:42:31.141Z\"}"

		ApiRoot := "https://apps.mypurecloud.ie"
		response, _ := apiClient.Post(ApiRoot+"/api/v2/analytics/conversations/details/query",
			"application/json",
			bytes.NewBuffer([]byte(query)))

		entities := &ConversationsResponse{}
		dec := json.NewDecoder(response.Body)
		dec.Decode(entities)

		if len(entities.Conversations) == 0 {
			return
		}

		wg := &sync.WaitGroup{}

		for _, conversation := range entities.Conversations {
			wg.Add(1)
			go func(id string) { //start a new goroutine
				disconnect := "{\"state\": \"disconnected\"}"
				request, err := http.NewRequest("PATCH",
					ApiRoot+"/api/v2/conversations/chats/"+id,
					bytes.NewBuffer([]byte(disconnect)))

				request.Header.Set("content-type", "application/json")
				deleteresponse, err := apiClient.Do(request)

				if err != nil {
					log.Panicf("can't disconnect %v", err)
				} else if deleteresponse.StatusCode == 429 {
					log.Print("SLEEPING")
					time.Sleep(45 * time.Second)
				} else if deleteresponse.StatusCode != 200 {
					log.Printf("Got unexpected response %v", deleteresponse.StatusCode)
				}
				wg.Done()
			}(conversation.ConversationID)
		}

		wg.Wait() // wait for all the disconnects to finish then get the next page
	}
}
