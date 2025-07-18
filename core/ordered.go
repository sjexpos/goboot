package core

const ORDERED_HIGHEST_PRECEDENCE = 0
const ORDERED_LOWEST_PRECEDENCE = int(^uint(0) >> 1)

type Ordered interface {
	GetOrder() int
}
