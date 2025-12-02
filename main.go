package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Data struct {
	SquadName  string            `json:"-" ido:"-"`
	HomeTown   string            `json:"homeTown"`
	Formed     int64             `json:"formed"`
	People     []string          `json:"people"`
	SecretBase string            `json:"secretBase"`
	Active     bool              `json:"active"`
	Heights    []int64           `json:"heights"`
	Dim        [][]ComputerParts `json:"dim"`
	Dim2       []ComputerParts   `json:"dim2"`
	Dim3       [][]int64         `json:"dim3"`
	Members    []Member          `json:"members"`
	Bank       Bank              `json:"bank"`
	CreatedAt  time.Time         `json:"created_at"`
}

type Member struct {
	Name           string          `json:"name"`
	Age            int             `json:"age"`
	SecretIdentity string          `json:"secretIdentity"`
	Powers         []string        `json:"powers"`
	ComputerParts  []ComputerParts `json:"computer_parts"`
}

type ComputerParts struct {
	Monitor  string `json:"monitor"`
	Mouse    string `json:"mouse"`
	Keyboard string `json:"keyboard"`
	Speakers string `json:"speakers"`
	GPU      string `json:"gpu"`
	CPU      string `json:"cpu"`
}

type Person struct {
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	Age      int    `json:"age"`
	Bank     Bank   `json:"bank"`
}

type Bank struct {
	Location string  `json:"location"`
	Money    float64 `json:"money"`
	Accounts int64   `json:"accounts"`
	Country  string  `json:"country"`
}

func test1() {
	obj := Data{}
	obj.CreatedAt = time.Now()
	obj.SquadName = `Super" hero "squad`
	obj.HomeTown = "Metro City"
	obj.Formed = 2016
	obj.People = []string{`one`, "two", "three"}
	obj.SecretBase = "Super tower"
	obj.Active = true
	obj.Heights = []int64{1, 2, 3, 4, 5}
	obj.Dim = [][]ComputerParts{[]ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}}, []ComputerParts{ComputerParts{Monitor: "Samsung"}}}
	obj.Dim2 = []ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}, ComputerParts{Monitor: "Samsung"}}
	obj.Dim3 = [][]int64{[]int64{5, 4, 3, 2, 1}, []int64{6, 7, 8, 9, 10}}
	obj.Members = append(obj.Members, Member{Name: "Molecule Man", Age: 29,
		SecretIdentity: "Dan Jukes", Powers: []string{"Radiation resistance", "Turning tiny", "Radiation blast"}})

	obj.Members[0].ComputerParts = append(obj.Members[0].ComputerParts, ComputerParts{
		Monitor: "Philips", Mouse: "Corsair", Keyboard: "The Cheapest", GPU: "RTX 2060", CPU: "AMD",
	})

	start := time.Now()
	data, _ := Marshal(obj)
	fmt.Printf("ido: %d, time: %s\n", len(data), time.Since(start))

	start = time.Now()
	j, _ := json.Marshal(obj)
	fmt.Printf("json: %d, time: %s\n", len(j), time.Since(start))

	start = time.Now()
	var obj1 Data
	Unmarshal(data, &obj1)
	fmt.Printf("ido unmarshal time: %s\n", time.Since(start))
}

func test2() {
	obj := Data{}
	obj.SquadName = `Super" hero "squad`
	obj.HomeTown = "Metro City"
	obj.Formed = 2016
	obj.People = []string{`one`, "two", "three"}
	obj.SecretBase = "Super tower"
	obj.Active = true
	obj.Heights = []int64{1, 2, 3, 4, 5}
	obj.Dim = [][]ComputerParts{[]ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}}, []ComputerParts{ComputerParts{Monitor: "Samsung"}}}
	obj.Dim2 = []ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}, ComputerParts{Monitor: "Samsung"}}
	obj.Dim3 = [][]int64{[]int64{5, 4, 3, 2, 1}, []int64{6, 7, 8, 9, 10}}
	obj.Members = append(obj.Members, Member{Name: "Molecule Man", Age: 29,
		SecretIdentity: "Dan Jukes", Powers: []string{"Radiation resistance", "Turning tiny", "Radiation blast"}})

	obj.Members[0].ComputerParts = append(obj.Members[0].ComputerParts, ComputerParts{
		Monitor: "Philips", Mouse: "Corsair", Keyboard: "The Cheapest", GPU: "RTX 2060", CPU: "AMD",
	})

	for i := 0; i < 300000; i++ {
		obj.Members[0].ComputerParts = append(obj.Members[0].ComputerParts, ComputerParts{
			Monitor: "Philips", Mouse: "Corsair", Keyboard: "The Cheapest", GPU: "RTX 2060", CPU: "AMD",
		})
	}

	start := time.Now()
	data, _ := Marshal(obj)
	fmt.Printf("ido: %d, time: %s\n", len(data), time.Since(start))

	start = time.Now()
	j, _ := json.Marshal(obj)
	fmt.Printf("json: %d, time: %s\n", len(j), time.Since(start))

	start = time.Now()
	var obj1 Data
	Unmarshal(data, &obj1)
	fmt.Printf("ido unmarshal time: %s\n", time.Since(start))

	start = time.Now()
	var obj3 Data
	json.Unmarshal(j, &obj3)
	fmt.Printf("json unmarshal time: %s\n", time.Since(start))
}

func test3() {
	obj := Data{}
	obj.SquadName = `Super" hero "squad`
	obj.HomeTown = "Metro City"
	obj.Formed = 2016
	obj.People = []string{`one`, "two", "three"}
	obj.SecretBase = "Super tower"
	obj.Active = true
	obj.Heights = []int64{1, 2, 3, 4, 5}
	obj.Dim = [][]ComputerParts{[]ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}}, []ComputerParts{ComputerParts{Monitor: "Samsung"}}}
	obj.Dim2 = []ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}, ComputerParts{Monitor: "Samsung"}}
	obj.Dim3 = [][]int64{[]int64{5, 4, 3, 2, 1}, []int64{6, 7, 8, 9, 10}}
	obj.Members = append(obj.Members, Member{Name: "Molecule Man", Age: 29,
		SecretIdentity: "Dan Jukes", Powers: []string{"Radiation resistance", "Turning tiny", "Radiation blast"}})

	obj.Members[0].ComputerParts = append(obj.Members[0].ComputerParts, ComputerParts{
		Monitor: "Philips", Mouse: "Corsair", Keyboard: "The Cheapest", GPU: "RTX 2060", CPU: "AMD",
	})

	for i := 0; i < 300000; i++ {
		obj.Members[0].ComputerParts = append(obj.Members[0].ComputerParts, ComputerParts{
			Monitor: "Philips", Mouse: "Corsair", Keyboard: "The Cheapest", GPU: "RTX 2060", CPU: "AMD",
		})
	}

	start1 := time.Now()
	data, _ := Marshal(obj)
	end1 := time.Since(start1)
	//fmt.Println(data)

	start2 := time.Now()
	jsonData, _ := json.Marshal(obj)
	end2 := time.Since(start2)
	//fmt.Println(string(jsonData))

	fmt.Printf("Marshal Json: %v characters in %v, dux: %v characters in %v\n", len(string(jsonData)), end2, len(data), end1)
}

func test4() {
	obj := Data{}
	obj.SquadName = `Super" hero "squad`
	obj.HomeTown = "Metro City"
	obj.Formed = 2016
	obj.People = []string{`one`, "two", "three"}
	obj.SecretBase = "Super tower"
	obj.Active = true
	obj.Heights = []int64{1, 2, 3, 4, 5}
	obj.Dim = [][]ComputerParts{[]ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}}, []ComputerParts{ComputerParts{Monitor: "Samsung"}}}
	obj.Dim2 = []ComputerParts{ComputerParts{Monitor: "Philips", Mouse: "test1"}, ComputerParts{Monitor: "Samsung"}}
	obj.Dim3 = [][]int64{[]int64{5, 4, 3, 2, 1}, []int64{6, 7, 8, 9, 10}}
	obj.Members = append(obj.Members, Member{Name: "Molecule Man", Age: 29,
		SecretIdentity: "Dan Jukes", Powers: []string{"Radiation resistance", "Turning tiny", "Radiation blast"}})

	obj.Members[0].ComputerParts = append(obj.Members[0].ComputerParts, ComputerParts{
		Monitor: "Philips", Mouse: "Corsair", Keyboard: "The Cheapest", GPU: "RTX 2060", CPU: "AMD",
	})

	obj.Bank = Bank{"Location 1", 340.40, 10000, "Spain"}

	/*data, _ := ido.MarshalOnly(obj, []byte("{1,2,3,4,11{0,2,3},8}"))
	fmt.Println(string(data))*/
}

func test5() {
	bank := Bank{
		Location: "Santander",
		Money:    10000000.0,
		Accounts: 100,
		Country:  "Spain",
	}

	person := Person{
		Name:     "John",
		LastName: "Doe",
		Age:      30,
		Bank:     bank,
	}

	data, _ := Marshal(person)
	fmt.Println(string(data))

	data2, _ := json.Marshal(person)
	fmt.Println(string(data2))

	newBank := Bank{}
	Unmarshal(data, &newBank)
}

func main() {
	test5()
}
