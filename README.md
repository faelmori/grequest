![logo](https://github.com/user-attachments/assets/3bef5ed3-a40b-4634-9de4-a2dd43d57f3d)


**A simple and lightweight HTTP client for Go, inspired by Requests (Python) and Guzzle (PHP).**

Grequests provides a declarative, chainable API to streamline HTTP requests in Go, supporting JSON manipulation, form submissions, cookies, file handling, authentication, and proxy configuration—all while remaining lightweight and dependency-free.

---

## **Features**

- **Lightweight & Efficient**: No third-party dependencies, built directly on `net/http`.
- **Flexible Request Handling**: Supports GET, POST, PUT, DELETE, and other HTTP methods.
- **JSON Parsing**: Convert responses into Go structs or string maps.
- **Header & Cookie Management**: Easily set, retrieve, and persist headers and cookies.
- **File Handling**: Upload and download files seamlessly.
- **Authentication**: Supports Basic, Bearer, and custom token authentication.
- **Proxy Support**: Configure custom proxy servers for HTTP requests.

---

## **Installation**

Install Grequests using Go modules:

```sh
go get github.com/lib4u/grequest
```

Or build from source:

```sh
git clone https://github.com/lib4u/grequest.git
cd grequest
go build -o grequest ./
./grequest --version
```

---

## **Usage**

### **Basic GET Request**

```go
req := app.Get("https://jsonplaceholder.typicode.com/todos/1").Do()
fmt.Println(req.Body().GetStrings())
```

### **Parsing JSON Response**

```go
type Todo struct {
    UserID    int    `json:"userId"`
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

var todo Todo
req := app.Get("https://jsonplaceholder.typicode.com/todos/1").Do()
err := req.Body().GetWithJsonStruct(&todo)
if err != nil {
    fmt.Println("Error decoding JSON")
}
fmt.Println(todo.Title)
```

### **POST Request with JSON Payload**

```go
data := LoginRequest{
    Username: "example",
    Password: "12345",
}
req := app.Post("https://example.site/login").Body().SetJson(data).Do()
fmt.Println(req.Status().GetCode())
```

### **Downloading a File**

```go
app.Get("https://example.com/image.png").Do().Body().SaveFile()
```

### **Multipart Form Submission**

```go
req := app.Post("https://example.site/form/")
req.Header().Set("Client", "number_1")
form := req.FormData().WithMultipart()
form.AddField("first_name", "John")
form.AddFile("photo", "my_photo.png")
form.Push()
req.Do()
```

### **Authenticated Requests**

```go
// Basic Authentication
app.Post("https://example.site/secret").Auth().SetBasic("user", "password").Do()

// Bearer Token Authentication
app.Post("https://example.site/secret").Auth().SetBearer("myToken").Do()
```

---

## **Contributing**

1. Fork the repository and clone it locally.
2. Create a new branch (`git checkout -b feature/branch-name`).
3. Make your changes and commit (`git commit -m "Description of changes"`).
4. Push your changes and create a pull request.

---

## **License**

Grequests is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

For more details, visit the [GitHub repository](https://github.com/lib4u/grequest).

---