package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"sync"

	_ "github.com/lib/pq"
)

type Character struct {
	Name        string
	Description string
	ImageURL    string
	ReleaseDate string
	Role        string
	Class       string
	Price       int
}

var characters = []Character{
	{
		Name:        "Натан",
		Description: "Натан — стрелок, который использует свои уникальные способности для нанесения урона на расстоянии и маневрирования по полю боя. Он может перемещаться в пространстве, избегая атак врагов, и наносить мощный урон с помощью своих навыков. Его способности позволяют ему быстро переключаться между атакой и уклонением, что делает его опасным противником.",
		ImageURL:    "https://i.pinimg.com/originals/34/a1/97/34a197fa001faeb4255fe67a899e113a.jpg",
		ReleaseDate: "2021-07-21",
		Role:        "Стрелок",
		Class:       "Магический",
		Price:       32000,
	},
	{
		Name:        "Алдос",
		Description: "Алдос — боец с неплохим уроном и повышенной живучестью. В его арсенале есть мощные навыки, которые помогут преследовать и уничтожать врагов в любом месте карты.",
		ImageURL:    "https://i.pinimg.com/736x/0b/67/15/0b6715d1177e799decd44430ae11ac97.jpg",
		ReleaseDate: "2016-07-14",
		Role:        "Убийца",
		Class:       "Физический",
		Price:       32000,
	},
	{
		Name:        "Лэйла",
		Description: "Лэйла — стрелок, который может наносить урон на расстоянии и имеет мощные навыки.",
		ImageURL:    "https://i.pinimg.com/736x/4e/76/68/4e7668c59b55f51addf771af44cfe834.jpg",
		ReleaseDate: "2016-07-14",
		Role:        "Стрелок",
		Class:       "Физический",
		Price:       32000,
	},
}

var userCharacters = make(map[string][]Character)
var mu sync.Mutex
var db *sql.DB

func init() {
	var err error
	connStr := "user=postgres dbname=go password=0000 host=localhost sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	// Создаем таблицу пользователей, если она не существует
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
        username TEXT NOT NULL PRIMARY KEY,
        password TEXT NOT NULL
    );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
   <!DOCTYPE html>
   <html>
   <head>
       <title>Персонажи MLBB</title>
   </head>
   <body>
       <h1>Персонажи MLBB</h1>
       <form method="GET" action="/">
           <input type="text" name="search" placeholder="Поиск персонажа">
           <input type="submit" value="Поиск">
       </form>
       <ul>
           {{range .}}
               <li>
                   <h2>{{.Name}}</h2>
                   <img src="{{.ImageURL}}" alt="{{.Name}}" style="width:100px;">
                   <p>{{.Description}}</p>
                   <p>Роль: {{.Role}}, Класс: {{.Class}}, Цена: {{.Price}}</p>
               </li>
           {{end}}
       </ul>
       <a href="/register">Регистрация</a> | <a href="/login">Вход</a>
   </body>
   </html>`

	search := r.URL.Query().Get("search")
	var filteredCharacters []Character

	// Добавляем стандартных персонажей
	for _, character := range characters {
		if search == "" || contains(character.Name, search) {
			filteredCharacters = append(filteredCharacters, character)
		}
	}

	// Добавляем персонажей пользователей
	mu.Lock()
	for _, userChars := range userCharacters {
		for _, character := range userChars {
			if search == "" || contains(character.Name, search) {
				filteredCharacters = append(filteredCharacters, character)
			}
		}
	}
	mu.Unlock()

	t, _ := template.New("webpage").Parse(tmpl)
	t.Execute(w, filteredCharacters)
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		mu.Lock()
		_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
		mu.Unlock()
		if err != nil {
			http.Error(w, "Ошибка регистрации", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tmpl := `
       <!DOCTYPE html>
   <html>
   <head>
       <title>Регистрация</title>
   </head>
   <body>
       <h1>Регистрация</h1>
       <form method="POST">
           <input type="text" name="username" placeholder="Имя пользователя" required>
           <input type="password" name="password" placeholder="Пароль" required>
           <input type="submit" value="Зарегистрироваться">
       </form>
       <a href="/">Назад</a>
   </body>
   </html>`
	t, _ := template.New("register").Parse(tmpl)
	t.Execute(w, nil)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		mu.Lock()
		var storedPassword string
		err := db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
		mu.Unlock()
		if err != nil || storedPassword != password {
			http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "username",
			Value: username,
			Path:  "/",
		})
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	tmpl := `
   <!DOCTYPE html>
   <html>
   <head>
       <title>Вход</title>
   </head>
   <body>
       <h1>Вход</h1>
       <form method="POST">
           <input type="text" name="username" placeholder="Имя пользователя" required>
           <input type="password" name="password" placeholder="Пароль" required>
           <input type="submit" value="Войти">
       </form>
       <a href="/">Назад</a>
   </body>
   </html>`
	t, _ := template.New("login").Parse(tmpl)
	t.Execute(w, nil)
}

func profilePage(w http.ResponseWriter, r *http.Request) {
	username, err := r.Cookie("username")
	if err != nil {
		http.Error(w, "Необходимо войти в систему", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		characterName := r.FormValue("name")
		mu.Lock()
		if characters, ok := userCharacters[username.Value]; ok {
			for i, character := range characters {
				if character.Name == characterName {
					userCharacters[username.Value] = append(characters[:i], characters[i+1:]...)
					break
				}
			}
		}
		mu.Unlock()
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	tmpl := `
   <!DOCTYPE html>
   <html>
   <head>
       <title>Профиль пользователя</title>
   </head>
   <body>
       <h1>Профиль: {{.Username}}</h1>
       <h2>Добавить персонажа</h2>
       <form method="POST" action="/addCharacter">
           <input type="text" name="name" placeholder="Имя" required>
           <input type="text" name="description" placeholder="Описание" required>
           <input type="text" name="imageURL" placeholder="URL изображения" required>
           <input type="text" name="releaseDate" placeholder="Дата выпуска" required>
           <input type="text" name="role" placeholder="Роль" required>
           <input type="text" name="class" placeholder="Класс" required>
           <input type="submit" value="Добавить">
       </form>
       <h2>Ваши персонажи</h2>
       <ul>
           {{range .Characters}}
               <li>
                   {{.Name}}
                   <form method="POST" style="display:inline;">
                       <input type="hidden" name="name" value="{{.Name}}">
                       <input type="submit" value="Удалить">
                   </form>
                   <form method="GET" action="/editCharacter">
                       <input type="hidden" name="name" value="{{.Name}}">
                                               <input type="submit" value="Редактировать">
                   </form>
               </li>
           {{end}}
       </ul>
       <a href="/">Назад</a>
   </body>
   </html>`

	mu.Lock()
	userChars := userCharacters[username.Value]
	mu.Unlock()
	data := struct {
		Username   string
		Characters []Character
	}{
		Username:   username.Value,
		Characters: userChars,
	}
	t, _ := template.New("profile").Parse(tmpl)
	t.Execute(w, data)
}

func editCharacterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		characterName := r.URL.Query().Get("name")
		var character Character
		mu.Lock()
		for _, chars := range userCharacters {
			for _, char := range chars {
				if char.Name == characterName {
					character = char
					break
				}
			}
		}
		mu.Unlock()

		tmpl := `
       <!DOCTYPE html>
       <html>
       <head>
           <title>Редактировать персонажа</title>
       </head>
       <body>
           <h1>Редактировать персонажа: {{.Name}}</h1>
           <form method="POST">
               <input type="hidden" name="oldName" value="{{.Name}}">
               <input type="text" name="name" value="{{.Name}}" required>
               <input type="text" name="description" value="{{.Description}}" required>
               <input type="text" name="imageURL" value="{{.ImageURL}}" required>
               <input type="text" name="releaseDate" value="{{.ReleaseDate}}" required>
               <input type="text" name="role" value="{{.Role}}" required>
               <input type="text" name="class" value="{{.Class}}" required>
               <input type="submit" value="Сохранить изменения">
           </form>
           <a href="/profile">Назад</a>
       </body>
       </html>`
		t, _ := template.New("editCharacter").Parse(tmpl)
		t.Execute(w, character)
		return
	}

	if r.Method == "POST" {
		oldName := r.FormValue("oldName")
		newCharacter := Character{
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			ImageURL:    r.FormValue("imageURL"),
			ReleaseDate: r.FormValue("releaseDate"),
			Role:        r.FormValue("role"),
			Class:       r.FormValue("class"),
			Price:       32000,
		}

		mu.Lock()
		for username, chars := range userCharacters {
			for i, char := range chars {
				if char.Name == oldName {
					userCharacters[username][i] = newCharacter
					break
				}
			}
		}
		mu.Unlock()

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
}

func addCharacterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username, err := r.Cookie("username")
		if err != nil {
			http.Error(w, "Необходимо войти в систему", http.StatusUnauthorized)
			return
		}
		newCharacter := Character{
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			ImageURL:    r.FormValue("imageURL"),
			ReleaseDate: r.FormValue("releaseDate"),
			Role:        r.FormValue("role"),
			Class:       r.FormValue("class"),
			Price:       32000,
		}
		mu.Lock()
		userCharacters[username.Value] = append(userCharacters[username.Value], newCharacter)
		mu.Unlock()
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func main() {
	defer db.Close() // Закрываем базу данных при завершении программы
	http.HandleFunc("/", homePage)
	http.HandleFunc("/register", registerPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/profile", profilePage)
	http.HandleFunc("/editCharacter", editCharacterPage)
	http.HandleFunc("/addCharacter", addCharacterPage)
	http.ListenAndServe(":8086", nil)
}
