package main

import (
	"html/template"
	"net/http"
	"sync"
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

var users = make(map[string]string)               // username: password
var userCharacters = make(map[string][]Character) // username: characters
var mu sync.Mutex

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

	// Добавляем стандартные персонажи
	for _, character := range characters {
		if search == "" || contains(character.Name, search) {
			filteredCharacters = append(filteredCharacters, character)
		}
	}

	// Добавляем персонажи пользователей
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
		users[username] = password
		mu.Unlock()
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
		if storedPassword, ok := users[username]; ok && storedPassword == password {
			http.SetCookie(w, &http.Cookie{
				Name:  "username",
				Value: username,
				Path:  "/",
			})
			mu.Unlock()
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}
		mu.Unlock()
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
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
					{{.Name}} <form method="POST" style="display:inline;">
					<input type="hidden" name="name" value="{{.Name}}">
					<input type="submit" value="Удалить">
					</form></li>
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
			Price:       32000, // Можно сделать динамическим
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
	http.HandleFunc("/", homePage)
	http.HandleFunc("/register", registerPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/profile", profilePage)
	http.HandleFunc("/addCharacter", addCharacterPage)
	http.ListenAndServe(":8083", nil)
}
