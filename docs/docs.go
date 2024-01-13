// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "get the status of server.",
                "tags": [
                    "root"
                ],
                "summary": "Show the status of server.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/v1/user/getToken": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Get new Tokens, also return ` + "`" + `refreshToken` + "`" + `",
                "tags": [
                    "User"
                ],
                "summary": "Get Token",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "return refreshToken in response body instead of saving in cookie",
                        "name": "noCookie",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "boolean",
                        "description": "also return profile images, slower response",
                        "name": "profileImages",
                        "in": "query",
                        "required": true
                    },
                    {
                        "description": "Device Info",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.DeviceInfo"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.UserViewModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/user/login": {
            "post": {
                "description": "Login with provided credentials",
                "tags": [
                    "User"
                ],
                "summary": "Login user",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "return refreshToken in response body instead of saving in cookie",
                        "name": "noCookie",
                        "in": "query",
                        "required": true
                    },
                    {
                        "description": "User object",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.LoginViewModel"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.UserViewModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/user/logout": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Logout user, return accessToken as empty string and also reset/remove refreshToken cookie if use in browser\n.in other environments reset refreshToken from client after successful logout.",
                "tags": [
                    "User"
                ],
                "summary": "Logout",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        },
        "/v1/user/signup": {
            "post": {
                "description": "Register a new user with the provided credentials\nUnlike the main server, this one doesn't handle ip detection and ip location\nAlso detect multiple login on same device as new device login, can be handled on client side with adding 'deviceInfo.fingerprint'\nAlso doesn't handle and send emails",
                "tags": [
                    "User"
                ],
                "summary": "Register a new user",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "return refreshToken in response body instead of saving in cookie",
                        "name": "noCookie",
                        "in": "query",
                        "required": true
                    },
                    {
                        "description": "User object",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.RegisterViewModel"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.UserViewModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/response.ResponseErrorModel"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.DeviceInfo": {
            "type": "object",
            "properties": {
                "appName": {
                    "type": "string"
                },
                "appVersion": {
                    "type": "string"
                },
                "deviceModel": {
                    "type": "string"
                },
                "fingerprint": {
                    "type": "string"
                },
                "os": {
                    "type": "string"
                }
            }
        },
        "model.LoginViewModel": {
            "type": "object",
            "properties": {
                "deviceInfo": {
                    "$ref": "#/definitions/model.DeviceInfo"
                },
                "password": {
                    "type": "string"
                },
                "username_email": {
                    "type": "string"
                }
            }
        },
        "model.ProfileImage": {
            "type": "object",
            "properties": {
                "addDate": {
                    "type": "string"
                },
                "originalSize": {
                    "type": "integer"
                },
                "size": {
                    "type": "integer"
                },
                "thumbnail": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                },
                "userId": {
                    "type": "integer"
                }
            }
        },
        "model.RegisterViewModel": {
            "type": "object",
            "properties": {
                "confirmPassword": {
                    "type": "string"
                },
                "deviceInfo": {
                    "$ref": "#/definitions/model.DeviceInfo"
                },
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "model.TokenViewModel": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "type": "string"
                },
                "accessToken_expire": {
                    "type": "integer"
                },
                "refreshToken": {
                    "type": "string"
                }
            }
        },
        "model.UserViewModel": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "profileImages": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.ProfileImage"
                    }
                },
                "token": {
                    "$ref": "#/definitions/model.TokenViewModel"
                },
                "userId": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "response.ResponseErrorModel": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "errorMessage": {}
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Type \"Bearer\" followed by a space and JWT token.",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "2.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Fiber Swagger Example API",
	Description:      "This is a sample server server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
