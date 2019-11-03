package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "strconv"
    "github.com/gorilla/mux"
    "sync"
)


var l sync.Mutex

type Article struct {
    Id   string `json:"id"`
    Title string `json:"title"`
    Date string `json:"date"`
    Body string `json:"body"`
    Tags []string `json:"tags"`
}

type TagSearchResult struct {

    Tag string `json:"tag"`
    Count int `json:"count"`
    Articles []string `json:"articles"`
    RelatedTags []string  `json:"related_tags"`
}

var storage []Article
var tagMap = make(map[string][]Article)


func getArticleById(w http.ResponseWriter, r *http.Request) {
    pathParams := mux.Vars(r)
    w.Header().Set("Content-Type", "application/json")

    articleId := -1
    if val, ok := pathParams["articleId"]; ok {
        articleIdv, err := strconv.Atoi(val)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte(`{"message": "need a number"}`))
            return
        }
        articleId = articleIdv
    }

    if(articleId == 0){
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"message": "need a number greater than 0"}`))
        return
    }

    articleId = articleId - 1

    if(articleId >= len(storage)){
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`{"message": "Record not found"}`))
        return
    }

    var d1 = storage[articleId]
    res1B, _ := json.Marshal(d1)
    w.Write([]byte(string(res1B)))
}

func getArticleByTagAndDate(w http.ResponseWriter, r *http.Request) {
    pathParams := mux.Vars(r)
    w.Header().Set("Content-Type", "application/json")

    if val, ok := pathParams["tagName"]; ok {
        var articleList = tagMap[val]

        if(len(articleList) == 0){
            w.WriteHeader(http.StatusNotFound)
            w.Write([]byte(`{"message": "Record not found"}`))
            return
        }
        var result = TagSearchResult{

        }

        result.Tag = val

        if val1, ok := pathParams["date"]; ok {

            var dateAstring string
            if(len(val1) != 8){
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(`{"message": "Invalid date format"}`))
                return
            }

            dateAstring = val1[0:4] + "-" + val1[4:6] + "-"  + val1[6:]
            var count = 0
            var tg = make(map[string]int)

            for i := 0; i < len(articleList); i++ {
                var a = articleList[i]

                if(a.Date == dateAstring){
                    count++
                    result.Articles = append(result.Articles, a.Id)

                    for e := range a.Tags {
                        tg[a.Tags[e]] = 1
                    }

                }
            }

            for e := range tg {
                result.RelatedTags = append(result.RelatedTags, e)
            }

            result.Count = count

            if(len(result.Articles) > 10){
                var start = len(result.Articles) - 10
                result.Articles = result.Articles[start:]
            }


            res1B, _ := json.Marshal(result)
            w.Write([]byte(string(res1B)))
            return
        }

    }


}


func createArticle(w http.ResponseWriter, r *http.Request) {
    var newArticle Article
    reqBody, err := ioutil.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"message": "create failed"}`))
        return
    }

    l.Lock()
    newArticle.Id = strconv.FormatInt(int64(len(storage) + 1), 10)
    json.Unmarshal(reqBody, &newArticle)
    storage = append(storage, newArticle)
    addTagMap(newArticle)
    l.Unlock()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(newArticle)
}

func addTagMap(article Article){

    for i := 0; i < len(article.Tags); i++ {
        tagMap[article.Tags[i]] = append(tagMap[article.Tags[i]], article)
    }

}

func main() {

    var d1 = Article{
        Id:   "1",
        Title : "latest science shows that potato chips are better for you than sugar",
        Date: "2016-09-22",
        Body: "some text, potentially containing simple markup about how potato chips are great",
        Tags: []string{"health", "fitness", "science"},
    }

    storage = append(storage, d1)
    addTagMap(d1)

    r := mux.NewRouter()

    r.HandleFunc("/articles/{articleId}", getArticleById).Methods(http.MethodGet)

    r.HandleFunc("/articles", createArticle).Methods(http.MethodPost)
    r.HandleFunc("/tag/{tagName}/{date}", getArticleByTagAndDate).Methods(http.MethodGet)

    log.Fatal(http.ListenAndServe(":8080", r))
}