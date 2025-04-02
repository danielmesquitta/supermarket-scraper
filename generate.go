package root

//go:generate prisma-go-tools entities --schema ./sql/schema.prisma --output ./internal/domain/entity
//go:generate prisma-go-tools tables --schema ./sql/schema.prisma --output ./internal/provider/db/sqlite/schema
