package main

import (
	"encoding/json"
	gabs "github.com/Jeffail/gabs/v2"
	gin "github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ldMode                string
	sesamAPI              string
	sesamJWT              string
	port                  string
	datasets              []string
	namespaceMappings     map[string]string
	jsonLdContext         *gabs.Container
	jsonLdServiceEndpoint string
)

func loadConfig() {
	// get config
	log.Println("Loading Config ---------------------- ")
	sesamAPI = os.Getenv("SESAM_API")
	sesamJWT = os.Getenv("SESAM_JWT")
	ldMode = os.Getenv("JSONLD_MODE") // one of CONTEXT_FIRST, CONTEXT_REF_INLINE
	port = os.Getenv("SERVICE_PORT")
	jsonLdServiceEndpoint = os.Getenv("JSON_LD_SERVICE_API")
	if port == "" {
		port = "5000"
	}
	var datasetNames = os.Getenv("SESAM_DATASETS")
	var splitNames = strings.Split(datasetNames, ";")
	for _, n := range splitNames {
		datasets = append(datasets, n)
	}

	log.Println("API: " + sesamAPI)
	log.Println("JWT: " + sesamJWT)
	log.Println("PORT: " + port)
	log.Println("Datasets: " + datasetNames)
	log.Println("Loaded Config  ---------------------- ")
}

func getJSON(path string) ([]byte, error) {
	var url = sesamAPI + path
	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	var req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+sesamJWT)

	var resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

func fetchNodeMetadata() error {
	jsonData, err := getJSON("/metadata")
	if err != nil {
		return err
	}
	namespaceMappings = make(map[string]string)
	jsonParsed, _ := gabs.ParseJSON(jsonData)
	nsMappings := jsonParsed.Path("config.effective.namespaces.default")
	jsonLdContext = gabs.New()
	for key, child := range nsMappings.ChildrenMap() {
		namespaceMappings[key] = child.Data().(string)
		jsonLdContext.Set(child.Data().(string), "@context", key)
	}
	return nil
}

func getEntitiesStream(dataset string, since string) (*io.ReadCloser, error) {
	var url = sesamAPI + "/datasets/" + dataset + "/entities"
	if since != "" {
		url += "?since=" + since
	}
	var client = &http.Client{}
	var req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+sesamJWT)
	var resp, _ = client.Do(req)
	return &resp.Body, nil
}

func writeJSONLdStream(reader *io.ReadCloser, writer *gin.ResponseWriter) {
	dec := json.NewDecoder(*reader)
	enc := json.NewEncoder(*writer)

	// eat the [
	dec.Token()

	// output start of array and context
	(*writer).Write([]byte("["))

	var writtenFirst = false

	if ldMode == "CONTEXT_FIRST" {
		enc.Encode(jsonLdContext.Data())
		writtenFirst = true
	}

	// output the entities
	for dec.More() {
		var v map[string]interface{}
		if err := dec.Decode(&v); err != nil {
			log.Println(err)
			return
		}

		if writtenFirst {
			(*writer).Write([]byte(","))
		}

		// get _id and add @id
		v["@id"] = v["_id"]
		if ldMode != "CONTEXT_FIRST" {
			v["@context"] = jsonLdServiceEndpoint + "/context"
		}

		if err := enc.Encode(&v); err != nil {
			log.Println(err)
		}

		writtenFirst = true
	}

	(*writer).Write([]byte("]"))
}

func contains(list []string, item string) bool {
	for _, a := range list {
		if a == item {
			return true
		}
	}
	return false
}

func main() {
	log.Println("Starting JSON-LD Publisher")
	loadConfig()
	var nodeMetadataFetchError = fetchNodeMetadata()
	if nodeMetadataFetchError != nil {
		log.Panic(nodeMetadataFetchError)
	}

	r := gin.Default()

	r.GET("/context", func(c *gin.Context) {
		log.Println("Request for context")
		writer := c.Writer
		enc := json.NewEncoder(writer)
		c.Status(http.StatusOK)
		enc.Encode(jsonLdContext.Data())
	})

	r.GET("/datasets/:datasetID", func(c *gin.Context) {
		since := c.Query("since")
		datasetID := c.Param("datasetID")
		log.Println("Request for dataset: " + datasetID + " since: " + since)
		if !contains(datasets, datasetID) {
			log.Println("Dataset is not publised " + datasetID)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		var reader, _ = getEntitiesStream(datasetID, since)
		defer (*reader).Close()
		writer := c.Writer

		c.Status(http.StatusOK)
		writeJSONLdStream(reader, &writer)
	})
	r.Run(":" + port)
}
