IDO (Invicta Dux Object) is a lightweight, human-readable file format designed for fast serialization and compact data representation.
It blends concepts from JSON and CSV, offering a structure that looks like JSON but stores values positionally instead of using key–value pairs.
This makes IDO significantly smaller in size and faster to marshal in many cases.

Unlike JSON, IDO does not store field names, only values making it ideal for internal systems, high-performance pipelines, and environments where schema is known ahead of time.

IDO provides a compact, efficient, human-readable alternative to JSON for environments where field names are unnecessary and schema is fixed. Its combination of readability, reduced size, and fast encoding makes it a useful format for specialized high-performance applications.

## Go Example

```go
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

func main() {
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

    data, _ := ido.Marshal(person)
    fmt.Println(string(data))

    data2, _ := json.Marshal(person)
    fmt.Println(string(data2))
}
```

#### JSON Output

```json
{
    "name": "John",
    "last_name": "Doe",
    "age": 30,
    "bank": {
        "location": "Santander",
        "money": 10000000,
        "accounts": 100,
        "country": "Spain"
    }
}
```

#### IDO Output

```ido
{
    "John",
    "Doe",
    30,
    {
        "Santander",
        10000000,
        100,
        "Spain"
    }
}
```
As you can see, IDO removes keys entirely and stores a compact, ordered data array.

### Performance Comparison

The following benchmark serializes and deserializes a large nested structure containing strings, arrays, slices, and deeply nested objects.

```go
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

func main() {
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
    data, _ := ido.Marshal(obj)
    fmt.Printf("ido: %d, time: %s\n", len(data), time.Since(start))

    start = time.Now()
    j, _ := json.Marshal(obj)
    fmt.Printf("json: %d, time: %s\n", len(j), time.Since(start))

    start = time.Now()
    var obj1 Data
    err := ido.Unmarshal(data, &obj1)
    fmt.Printf("ido unmarshal time: %s, err: %v\n", time.Since(start), err)

    start = time.Now()
    var obj3 Data
    err = json.Unmarshal(j, &obj3)
    fmt.Printf("json unmarshal time: %s, err: %v\n", time.Since(start), err)
}
```

#### Result

```
ido: 165000337, time: 620.132852ms
json: 327000924, time: 1.216462615s
ido unmarshal time: 2.66366359s, err: <nil>
json unmarshal time: 5.228011242s, err: <nil>
```