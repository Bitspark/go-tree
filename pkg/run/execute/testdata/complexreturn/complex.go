// Package complexreturn provides functions that return complex types for testing
package complexreturn

// Person represents a person
type Person struct {
	Name    string
	Age     int
	Address Address
}

// Address represents a postal address
type Address struct {
	Street  string
	City    string
	Country string
	Zip     string
}

// GetPerson returns a person object
func GetPerson(name string) Person {
	return Person{
		Name: name,
		Age:  30,
		Address: Address{
			Street:  "123 Main St",
			City:    "Anytown",
			Country: "USA",
			Zip:     "12345",
		},
	}
}

// GetPersonMap returns a map with person data
func GetPersonMap(name string) map[string]interface{} {
	return map[string]interface{}{
		"Name": name,
		"Age":  30,
		"Address": map[string]string{
			"Street":  "123 Main St",
			"City":    "Anytown",
			"Country": "USA",
			"Zip":     "12345",
		},
	}
}

// GetPersonSlice returns a slice of persons
func GetPersonSlice(names ...string) []Person {
	result := make([]Person, 0, len(names))
	for i, name := range names {
		result = append(result, Person{
			Name: name,
			Age:  30 + i,
			Address: Address{
				Street:  "123 Main St",
				City:    "Anytown",
				Country: "USA",
				Zip:     "12345",
			},
		})
	}
	return result
}

// ComplexStruct combines multiple types
type ComplexStruct struct {
	People   []Person
	Counts   map[string]int
	Active   bool
	Priority float64
	Tags     []string
	Metadata map[string]interface{}
}

// GetComplexStruct returns a complex struct with nested types
func GetComplexStruct() ComplexStruct {
	return ComplexStruct{
		People: []Person{
			{
				Name: "Alice",
				Age:  30,
				Address: Address{
					Street:  "123 Main St",
					City:    "Anytown",
					Country: "USA",
					Zip:     "12345",
				},
			},
			{
				Name: "Bob",
				Age:  35,
				Address: Address{
					Street:  "456 Oak Ave",
					City:    "Othertown",
					Country: "USA",
					Zip:     "67890",
				},
			},
		},
		Counts: map[string]int{
			"visits": 10,
			"clicks": 25,
			"views":  100,
		},
		Active:   true,
		Priority: 0.75,
		Tags:     []string{"important", "customer", "verified"},
		Metadata: map[string]interface{}{
			"created":  "2023-01-15",
			"modified": "2023-06-20",
			"status":   "active",
			"scores":   []int{85, 92, 78},
		},
	}
}
