package id

import (
	"github.com/bwmarrin/snowflake"
	"log"
)

var node *snowflake.Node

func init() {
	created, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalln(err)
	}
	node = created
}

// Generate generated a new snowflake ID
func Generate() snowflake.ID {
	return node.Generate()
}
