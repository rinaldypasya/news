package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/olivere/elastic"
	"github.com/vmihailenco/msgpack"

	"github.com/rinaldypasya/news/ampq/consumer"
	"github.com/rinaldypasya/news/ampq/producer"
	"github.com/rinaldypasya/news/news"
)

type SearchResponse struct {
	Time string      `json:"time"`
	Hits string      `json:"hits"`
	News []news.News `json:"news"`
}

const (
	elasticIndexName = "newsx"
	elasticTypeName  = "news"
)

var (
	elasticClient *elastic.Client
)

func (idb *InDB) PostNews(c *gin.Context) {
	var news news.News

	author := c.PostForm("author")
	body := c.PostForm("body")
	news.Author = author
	news.Body = body

	producer.PublishMessages(1, news)
	consumer.ConsumeMessages(news)
	idb.DB.Create(&news)
	idb.DB.First(&news)

	if err := c.BindJSON(&news); err != nil {
		errorResponse(c, http.StatusBadRequest, "Malformed request body")
		return
	}
	// Insert documents in bulk
	bulk := elasticClient.
		Bulk().
		Index(elasticIndexName).
		Type(elasticTypeName)

	bulk.Add(elastic.NewBulkIndexRequest().Id(strconv.Itoa(news.ID)).Doc(news))
	if _, err := bulk.Do(c.Request.Context()); err != nil {
		log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to create documents")
		return
	}
	c.JSON(http.StatusOK, news)
}

func (idb *InDB) GetNews(c *gin.Context) {
	// Parse request
	query := c.Query("query")
	if query == "" {
		errorResponse(c, http.StatusBadRequest, "Query not specified")
		return
	}
	skip := 0
	take := 10
	if i, err := strconv.Atoi(c.Query("skip")); err == nil {
		skip = i
	}
	if i, err := strconv.Atoi(c.Query("take")); err == nil {
		take = i
	}
	// Perform search
	esQuery := elastic.NewMultiMatchQuery(query, "id", "created").
		Fuzziness("2").
		MinimumShouldMatch("2")
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(esQuery).
		From(skip).Size(take).
		Sort("created", false).
		Do(c.Request.Context())
	if err != nil {
		log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		return
	}
	res := SearchResponse{
		Time: fmt.Sprintf("%v", result.TookInMillis),
		Hits: fmt.Sprintf("%v", result.Hits.TotalHits),
	}
	// Transform search results before returning them
	newsx := make([]news.News, 0)
	for _, hit := range result.Hits.Hits {
		var news news.News
		json.Unmarshal(hit.Source, &news)
		newsx = append(newsx, news)
	}
	go func() {
		idb.DB.Find(&newsx)
	}()

	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": ":6379",
		},
	})

	codec := &cache.Codec{
		Redis: ring,

		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	for news := range newsx {
		key := "newskey"
		codec.Set(&cache.Item{
			Key:    key,
			Object: news,
		})
	}

	res.News = newsx
	c.JSON(http.StatusOK, res)
}

func errorResponse(c *gin.Context, code int, err string) {
	c.JSON(code, gin.H{
		"error": err,
	})
}
