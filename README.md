# Go Chat

the chat and user management service of the [downloader_api](https://github.com/ashkan-esz/downloader_api) project.

## Motivation

Needed the chat system and managing user related logic with faster response and leave the main server handle crawling and movie related logics, so why not using golang?

## How to use


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

| Prop                                   | Description                                                                              | Required | Default Value |
|----------------------------------------|------------------------------------------------------------------------------------------|----------|---------------|
| **`PORT`**                             | server port                                                                              | `false`  | 3000          |
| **`POSTGRES_DATABASE_URL`**            |                                                                                          | `true`   |               |
| **`MONGODB_DATABASE_NAME`**            |                                                                                          | `true`   |               |
| **`MONGODB_DATABASE_URL`**             |                                                                                          | `true`   |               |
| **`REDIS_URL`**                        |                                                                                          | `true`   |               |
| **`REDIS_PASSWORD`**                   |                                                                                          | `true`   |               |
| **`SENTRY_DNS`**                       | see [sentry.io](https://sentry.io)                                                       | `false`  |               |
| **`SENTRY_RELEASE`**                   | see [sentry release](https://docs.sentry.io/product/releases/.)                          | `false`  |               |
| **`ACCESS_TOKEN_SECRET`**              |                                                                                          | `true`   |               |
| **`REFRESH_TOKEN_SECRET`**             |                                                                                          | `true`   |               |
| **`ACCESS_TOKEN_EXPIRE_HOUR`**         |                                                                                          | `true`   |               |
| **`REFRESH_TOKEN_EXPIRE_DAY`**         |                                                                                          | `true`   |               |
| **`WAIT_REDIS_CONNECTION_SEC`**        |                                                                                          | `true`   |               |
| **`ACTIVE_SESSIONS_LIMIT`**            | number of device that each account can have logged it                                    | `true`   | 5             |
| **`CORS_ALLOWED_ORIGINS`**             | address joined by `---` example: https://download-admin.com---https:download-website.com | `false`  |               |
| **`CLOUAD_STORAGE_ENDPOINT`**          | s3 sever url, for example see [arvancloud.com](https://www.arvancloud.com/en)            | `true`   |               |
| **`CLOUAD_STORAGE_WEBSITE_ENDPOINT`**  | s3 static website postfix                                                                | `true`   |               |
| **`CLOUAD_STORAGE_ACCESS_KEY`**        |                                                                                          | `true`   |               |
| **`CLOUAD_STORAGE_SECRET_ACCESS_KEY`** |                                                                                          | `true`   |               |
| **`BUCKET_NAME_PREFIX`**               | if bucket names not exist use this. for example 'poster' --> 'test_poster'               | `false`  |               |
| **`FIREBASE_AUTH_KEY`**                | a coded key from firebase that used in sending push notification                         | `true`   |               |
| **`RABBITMQ_URL`**                     |                                                                                          | `true`   |               |
| **`SERVER_ADDRESS`**                   | the url of the server                                                                    | `true`   |               |
| **`MAIN_SERVER_ADDRESS`**              | the url of the downloader_api (main server)                                              | `true`   |               |
| **`AGENDA_JOBS_COLLECTION`**           |                                                                                          | `true`   |               |
| **`DEFAULT_PROFILE_IMAGE`**            |                                                                                          | `true`   |               |
| **`MIGRATE_ON_START`**                 |                                                                                          | `true`   |               |
| **`PRINT_ERRORS`**                     |                                                                                          | `false`  | false         |
| **`DOMAIN`**                           | base domain, used for cookies domain and subdomain                                       | `true`   |               |
| **`BLURHASH_CONSUMER_COUNT`**          | number of parallel creation of blurHash                                                  | `false`  | 1             |
| **`APP_DEEP_LINK`**                    | deeplink of the mobile app, used in push notification                                    | `false`  |               |

>**NOTE: check [configs schema](https://github.com/ashkan-esz/downloader_api/blob/master/readme/CONFIGS.README.md) for other configs that read from db.**

## Future updates

- [x]  Fast and light.
- [ ]  Documentation.
- [ ]  Write test.

## Contributing

Contributions are always welcome!

See `contributing.md` for ways to get started.

Please adhere to this project's `code of conduct`.

## Support

Contributions, issues, and feature requests are welcome!
Give a ⭐️ if you like this project!

## Related

- [downloader_api](https://github.com/ashkan-esz/downloader_api)
- [downloader_app](https://github.com/ashkan-esz/downloader_app)

## Author

**Ashkan Esz**

- [Profile](https://github.com/ashkan-esz "Ashkan esz")
- [Email](mailto:ashkanaz2828@gmail.com?subject=Hi "Hi!")
