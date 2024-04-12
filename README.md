# Тестовое задание для стажёра Backend

## Описание задачи 
Необходимо реализовать сервис, который позволяет показывать пользователям баннеры, в зависимости от требуемой фичи и тега пользователя, а также управлять баннерами и связанными с ними тегами и фичами. 

Более полная информация о задаче доступна по сссылке <b>[avito-tech.](https://github.com/avito-tech/backend-trainee-assignment-2024)</b>

## Проблемы реализации 

1. Первая проблема возникла на этапе архитектуры, какую структуру использовать? Так как сервис достаточно большой решил использовать "Package based project layout" 
2. Вторая проблема связана с механизмом кэширования, так как не было запрета на использование сторонних сервисов, решил использовать redis
3. Третья и самая сложная проблема появилась на этапе тестирования, решил использовать вторую, моковую базу данных, которую также развернул в докере, сложность заключалась в том, что так как в основной базе данных я использовал pgx driver, пакет testing не поддерживал её и требовал driver "github.com/lib/pq"
4. В остальном были маленькие проблемы, которые преодолевались поиском информации в так называемом Google

## Инструкция по запуску

- Для первого запуска необходимо проинициализировать проект и базу данных через команду `make docker-build`
- Затем запустить приложение через команду `make up`
- В дальнейшем работа с приложением ведётся через команды `make up` для старта сервера и `make restart` для рестарта сервера 
- для запуска тестов использовать команду `make test`


# Описание эндпоинтов

## getUserBanner

This endpoint retrieves the user banner based on the provided feature ID and tag ID. Header can be either adminToken or userToken.

### Request Parameters

- `feature_id` (required query parameter) - The ID of the feature.
- `tag_id` (required query parameter) - The ID of the tag.
- `use_last_revision` (optional query parameter) - boolean set true to get latest banner information.

### Response

Upon a successful request, the server returns a status code of 200.
    
The response body contains an unstructured json object representing user banner. The values for these fields may vary based on the feature and tag IDs provided in the request.

#### cURL

curl --location 'http://localhost:8080/user_banner?feature_id=1&tag_id=3&use_last_revision=true' \
--header 'token: adminToken'


## getBanner

This endpoint makes an HTTP GET request to retrieve a banner with the specified parameters. 

- `feature_id` (optional query parameter) - The ID of the feature.
- `tag_id` (optional query parameter) - The ID of the tag.
- `limit` (optional query parameter) - Limit of the response.
- `offset` (optional query parameter) - Offset of the response

### Response

#### Response Example:
<pre>
    [{
        "id":1,
        "feature_id":1,
        "content":{"text":"some_test","title":"some_test","url":"some_test_url"},
        "is_active":false,
        "tag_ids":[1,2,3],
        "created_at":"2024-04-12T13:08:01.269634Z",
        "updated_at":"2024-04-12T13:08:01.269634Z"
    }]
</pre>

Upon a successful request, the server returns a status code of 200.


#### cURL

curl --location 'http://localhost:8080/banner?limit=1&offset=1' \
--header 'token: adminToken'



## createBanner

This endpoint allows you to create a new banner by sending an HTTP POST request to the specified URL.

#### Request Body

- `tag_ids` (array of integers) - An array of tag IDs associated with the banner.
- `feature_id` (integer) - The ID of the feature related to the banner.
- `content` (object) - An unstructured json object
- `is_active` (boolean) - Indicates whether the banner is active or not.
    

#### Response

Upon successful creation of the banner, the server will respond with StatusCode 201.

#### cURL

curl --location 'http://localhost:8080/banner' \
--header 'token: adminToken' \
--header 'Content-Type: application/json' \
--data '{
    "tag_ids": [1,2,3],
    "feature_id": 1,
    "content": {"title": "some_test", "text": "some_test", "url": "some_test_url"},
    "is_active": false
}'



## updateBanner

This endpoint is used to update a banner using an HTTP PATCH request. The request should be sent to [http://localhost:8080/banner](http://localhost:8080/banner) with a payload in raw format. The payload may or may not include tag_ids, feature_id, content (consisting of title, text, and URL), and a boolean value for is_active.

### Request Body

- tag_ids (array of integers): An array of tag IDs associated with the banner.
- feature_id (integer): The feature ID of the banner.
- content (object): An unstructured json object
- is_active (boolean): A boolean value indicating whether the banner is active.
    

### Response

The response to the request has a status code of 200

#### cURL

curl --location --request PATCH 'http://localhost:8080/banner' \
--header 'token: adminToken' \
--header 'Content-Type: application/json' \
--data '{
    "tag_ids": [3],
    "feature_id": 3,
    "content": {"title": "321", "text": "123", "url": "123"},
    "is_active": false
}'

### deleteBanner

This endpoint sends an HTTP DELETE request to remove the banner with the ID specified in url param.

### Response

Upon successful execution, the server returns a response with a status code of 204.

#### cURL

curl --location --request DELETE 'http://localhost:8080/banner/1' \
--header 'token: adminToken'
