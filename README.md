# Triangula API server

Minimalistic API server that calculates and serves artistic images using triangula

## Install

Download repository:

```
git clone github.com/maikschneider/triangula-api-server
```

Create `.env` file:

```
cp .env.example .env
```

Run:

```
go run server.go
```

## Usage

There are 3 Endpoints:

- `GET /` List all available images
- `POST /` Add new image for processing
- `GET /x23bxg2...` Download image

### List

Example response:

```
[
	{
		"Name": "fVV2b_gJlQt_v8Z6pSqPiw.svg",
		"Hash": "fVV2b_gJlQt_v8Z6pSqPiw",
		"Processed": true,
		"CallbackUrl": "https://example.com/myscript.php",
		"CreatedAt": 1641653220,
		"Settings": {}
	},
	{
		"Name": "eJ5nnzw2xMZqmG8zWkPLzA.svg",
		"Hash": "eJ5nnzw2xMZqmG8zWkPLzA",
		"Processed": true,
		"CallbackUrl": "https://example.com/myscript.php",
		"CreatedAt": 1641653265,
		"Settings": {}
	},
]
```

### Add

Expects a `multipart/form-data` post. Fieldnames:

| Field name             | required | description                                                                                   |
| ---------------------- | -------- | --------------------------------------------------------------------------------------------- |
| `file`                 | yes      | The image to process                                                                          |
| `callbackUrl`          | no       | The url that should be notified after successfull processing                                  |
| `hash`                 | no       | The md5 hash of the image. This can speed up the process since duplicate files are recognized |
| `settings[points]`     | no       | 300                                                                                           |
| `settings[shape]`      | no       | "triangles"                                                                                   |
| `settings[mutations]`  | no       | 2                                                                                             |
| `settings[variation]`  | no       | 0.3                                                                                           |
| `settings[population]` | no       | 400                                                                                           |
| `settings[cache]`      | no       | 22                                                                                            |
| `settings[block]`      | no       | 5                                                                                             |
| `settings[cutoff]`     | no       | 1                                                                                             |
| `settings[reps]`       | no       | 100                                                                                           |
| `settings[threads]`    | no       | 0                                                                                             |

Example response:

```
{
	"hash": "ZGDvaGMO54jx1oENQsSumQ",
	"message": "File queued"
}
```

### Download

To get the svg, just query the image using the hash:

```
GET /ZGDvaGMO54jx1oENQsSumQ
```


## Security

Add the api key defined in the `.env` file to every request. The header name is `X-API-KEY` HTTP header in every request

```
> GET / HTTP/1.1
> Host: 127.0.0.1:8080
> X-API-KEY: D78i1FbsOTOFRyhMjEoa
> Accept: */*

< HTTP/1.1 200 OK
```

## Clean up

The environment variable ```EXPIRATION``` defines how long images are kept in seconds.