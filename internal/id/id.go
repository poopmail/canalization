package id

import (
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

var node *snowflake.Node

func init() {
	created, err := snowflake.NewNode(1)
	if err != nil {
		logrus.WithError(err).Fatal()
	}
	node = created
}

// Generate generated a new snowflake ID
func Generate() snowflake.ID {
	return node.Generate()
}
