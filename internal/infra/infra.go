package infra

import (
	"github.com/xyzbit/ino/internal/infra/mysql"
	"github.com/xyzbit/ino/internal/infra/redis"
)

// Init 初始化所有基础设施
func Init() {
	mysql.Init()
	redis.Init()
	// TODO: 后续添加 InitMilvus() 和 InitNeo4j()
}
