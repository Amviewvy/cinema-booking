package models

type Show struct {
	ID    string `bson:"id" json:"id"`
	Movie string `bson:"movie" json:"movie"`
	Time  string `bson:"time" json:"time"`
	Price int    `bson:"price" json:"price"`
}
