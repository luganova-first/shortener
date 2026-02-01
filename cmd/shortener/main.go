package main

import (
    "io"
    "net/http"
    "crypto/rand"
)

// Хранилище сокращений, ключ -- хеш сокращения, значение -- сокращаемый URL
type Shorted map[string]string
var shorts Shorted

// Генератор хеша сокращения
func cryptoRandomString(length int) (string, error) {
    const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }

    for i := range bytes {
        bytes[i] = charset[bytes[i]%byte(len(charset))]
    }

    return string(bytes), nil
}

// Хендлер основной страницы
func mainPage(res http.ResponseWriter, req *http.Request) {
    // Проверяем методы запросов, должны быть только POST и GET
    if req.Method != http.MethodGet && req.Method != http.MethodPost {
        res.WriteHeader(http.StatusBadRequest)
        return
    }

    // POST запрос должен быть с Content-Type `text/plain`
    if req.Method == http.MethodPost && req.Header.Get("Content-Type") != "text/plain" {
        res.WriteHeader(http.StatusBadRequest)
        return
    }

    // Проверяем URL для POST запроса
    if req.Method == http.MethodPost && req.URL.Path != "/" {
        res.WriteHeader(http.StatusBadRequest)
        return
    }

    // Если POST запрос, читаем параметры запроса, в параметрах должен быть URL для сокращения
    if req.Method == http.MethodPost {
        defer req.Body.Close()
        body, err := io.ReadAll(req.Body)
        if err != nil {
            res.WriteHeader(http.StatusBadRequest)
            return
        }

        if string(body) == "" {
            res.WriteHeader(http.StatusBadRequest)
            return
        }

        targetValue := string(body)

        // Ищем body запроса в значениях уже сокращённых
        var cryptoString string
        for key, value := range shorts {
            if value == targetValue {
                cryptoString = key
            }
        }

        if cryptoString == "" {
            // Если не нашли, генерируем новое сокращение и записываем его в shorts
            str, err := cryptoRandomString(8)
            if err != nil {
                panic(err)
            }

            cryptoString = "/"+str

            shorts[cryptoString] = targetValue
        }

        res.Header().Set("content-type", "text/plain")
        res.WriteHeader(http.StatusCreated)
        res.Write([]byte("http://"+req.Host+cryptoString))
        return
    }

    // GET запрос, пытаемся найти сокращение в shorts по ключу
    if shorts[req.URL.Path] == "" {
        res.WriteHeader(http.StatusBadRequest)
        return
    }

    res.WriteHeader(http.StatusTemporaryRedirect)
    res.Write([]byte(shorts[req.URL.Path]))
    return
}

func main() {
    shorts = make(Shorted)

    mux := http.NewServeMux()
    mux.HandleFunc(`/`, mainPage)

    err := http.ListenAndServe(`:8080`, mux)
    if err != nil {
        panic(err)
    }
}
