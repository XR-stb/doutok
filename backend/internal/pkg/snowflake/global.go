package snowflake

var defaultGen *Generator

// Init 初始化全局 Snowflake 生成器
func Init(datacenterID, workerID int64) {
	// 将 datacenter 和 worker 合并为一个 10 bit 的 workerID
	combined := (datacenterID << 5) | workerID
	defaultGen = New(combined)
}

// GenID 使用全局生成器生成 ID
func GenID() int64 {
	if defaultGen == nil {
		defaultGen = New(1)
	}
	return defaultGen.Next()
}
