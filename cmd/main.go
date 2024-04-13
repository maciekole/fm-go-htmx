package main

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

var id int = 0

type Contact struct {
	Id    int
	Name  string
	Email string
}

func newContact(name string, email string) Contact {
	id++
	return Contact{
		Id:    id,
		Name:  name,
		Email: email,
	}
}

type Contacts = []Contact

func (d *Data) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

type Data struct {
	Contacts Contacts
}

func (d *Data) IndexOf(id int) int {
	for i, contact := range d.Contacts {
		if id == contact.Id {
			return i
		}
	}
	return -1
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("John", "jd@gmail.com"),
			newContact("Clara", "cd@gmail.com"),
		},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Count struct {
	Count int
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	page := newPage()
	e.Renderer = newTemplate()

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		fmt.Printf("DEBUG: %s", email)
		fmt.Println()

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email

			formData.Errors["email"] = "Email already exists!"
			return c.Render(422, "form", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)
		c.Render(200, "form", newFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(400, "Invalid ID")
		}

		index := page.Data.IndexOf(id)
		if index == -1 {
			return c.String(400, "Contact not found")
		}

		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)
		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
