# url-shortner-go

Make sure docker-compose is installed and docker daemon is running

## To start the service
```
docker-compose up -d
```

This will build the image for DB and the api service

If it's successful, it will give this output

> Creating url-shortner-go_db_1 ... <span style="color:green">done</span>
> Creating url-shortner-go_api_1 ... <span style="color:green">done</span>



## To check the data in redis

1. exec into the container
```
docker exec -it XYZ /bin/sh
```

2. Use the redis-cli and connect to the configured port
```
/data # redis-cli -p 6380
```


### To check quota for each IP address

3. Select index (from 0 or 1 for our usecase)
```
127.0.0.1:6380> select 1
OK
```

4. Check all the keys (IP addresses)
```
127.0.0.1:6380[1]> keys *
"172.30.0.1"
```

5. Get the value for it (ie the quota left for that IP address)
```
127.0.0.1:6380[1]> get "172.30.0.1"
"0"
```

### To check the mapping for CustomShort and the actual link

3. Select index (from 0 or 1 for our usecae)
```
127.0.0.1:6380> select 0
OK
```

4. Check all the keys (Custom Shorts)
```
127.0.0.1:6380> keys *
    ( 1) "904db9"
    ( 2) "a3e74f"
    ( 3) "8a67a4"
    ( 4) "0706af"
    ( 5) "3bbbbf"
    ( 6) "576437"
    ( 7) "3f9225"
    ( 8) "9b569d"
    ( 9) "42851d"
    (10) "cef11c"
```

5) Check value for a key
```
127.0.0.1:6380> get "904db9"
"https://www.youtube.com/watch?v=3ExDEeSnvE"
```


## To stop the server
```
docker-compose stop
```

## To delete the images
```
docker rmi url-shortner-go_api url-shortner-go_db
```

## To test the API
### Request
```
curl --location 'http://localhost:3000/api/v1' \
--header 'Content-Type: application/json' \
--data '{
    "url" : "https://www.youtube.com/watch?v=3ExEeSnvE"
}'
```

### Response
#### Valid
```
{
    "url": "https://www.youtube.com/watch?v=3ExEeSnvE",
    "short": "localhost:3000/681ffe",
    "expiry": 24,
    "rate_limit": 0,
    "rate_limit_reset": 0
}
```
#### Rate limit exceeded
```
{
    "error": "rate limit exceeded",
    "rate_limit_reset": 0
}
```
