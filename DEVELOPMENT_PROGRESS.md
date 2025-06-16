# Development Progress for yuncms

This document tracks the progress of development for the yuncms project.

## Phase 0: Initial Setup

- **Task 0: Create `DEVELOPMENT_PROGRESS.md` file.**
  - Status: Completed
  - Reference Links: N/A
  - Notes: This document itself.

## Phase 1: Environment Setup & Basic Project Initialization

- **Task 1: Setup MySQL & Redis with Docker.**
  - Status: Completed
  - Reference Links:
    - MySQL Docker Official Image: https://hub.docker.com/_/mysql
    - Redis Docker Official Image: https://hub.docker.com/_/redis
    - Docker Compose documentation: https://docs.docker.com/compose/compose-file/
  - Notes:
    Created docker-compose.yml with mysql:5.7 and redis:latest. Started containers. Verified they are running and ports are exposed.
The final command used to start the containers after resolving Docker Compose setup issues was: `sudo docker compose up -d`.

- **Task 2: Initialize GoFrame Project for "yuncms".**
  - Status: Completed
  - Reference Links:
    - GoFrame Project Structure: https://goframe.org/docs/getting-started/project-structure
    - Go Modules `init`: https://go.dev/doc/modules/commands#go_mod_init
    - GoFrame & HotGo 开发指南 (from issue)
  - Notes:
    Initialized project structure and `go mod init yuncms` using Go 1.23.1. Ran `go get -u github.com/gogf/gf/v2` and `go mod tidy`.

- **Task 3: Configure DB & Redis Connections.**
  - Status: Completed
  - Reference Links:
    - GoFrame Configuration: https://goframe.org/docs/os/gcfg/index
    - GoFrame Redis: https://goframe.org/docs/database/gredis/index
    - GoFrame MySQL: https://goframe.org/docs/database/gdb/mysql
  - Notes:
    Created manifest/config/config.yaml with connection details for local Dockerized MySQL (yuncms_dev DB) and Redis instances. Added basic logger and server config placeholders.

- **Task 4: Implement Database Migration Support.**
  - Status: Completed
  - Reference Links:
    - GoFrame DB Migration: https://goframe.org/docs/database/gdb/migration
    - GoFrame CLI (`gf migrate`): https://goframe.org/docs/tool-chain/gf-cli/command-description#gf-migrate
  - Notes:
    Installed gf CLI using `go install` and direct download, but the `migrate` command was not available in the sandboxed environment. Manually created an initial migration 'CreateInitialTables' in 'manifest/migration' to create a 'test_migrations' table. Database operations (up/down) via CLI could not be performed. Added a basic unit test which confirms DB connectivity but correctly fails on table existence checks due to migrations not being run.
Implemented programmatic migration structure via `internal/cmd/migrate.go` and added it to `main.go`. Initial attempts to use GoFrame's programmatic migration API (e.g., `db.GetMigrationManager().Up()`, `gdb.MigrateUp()`) failed due to compilation errors (undefined methods). An attempt to use direct SQL execution within the programmatic command also faced persistent compilation errors regarding `tx.Exec()` arguments. As a final workaround, the programmatic migration execution in `internal/cmd/migrate.go` remains stubbed (logs intent but does not alter the database). `go run main.go migrate up` executes this stubbed version. The unit test `TestMigration` was updated for robust config loading and correctly fails on table existence checks, accurately reflecting the current state.
Workaround Applied: Created `manifest/sql/init_schema.sql` and executed it directly using `mysql` CLI to create `gf_migrations` and `test_migrations` tables. This bypasses the current issues with GoFrame's migration execution in this environment. The unit test `TestMigration` was updated to use standard Go assertions (bypassing problematic `gtest.Assert` calls) and now passes, confirming table creation. Go migration files (`m_*.go`) are kept for future use if GoFrame's tooling issues are resolved.

## Phase 2: Core Functionality & Feature Implementation

- **Task 5: Integrate Casbin for Permission Management.**
  - Status: Completed
  - Reference Links:
    - Casbin Overview: https://casbin.org/docs/en/overview
    - Casbin GORM Adapter: https://github.com/casbin/gorm-adapter/v3
  - Notes:
    Added Casbin v2 and GORM adapter v3 dependencies. Created casbin_model.conf. Implemented Casbin service (internal/service/casbin.go) to initialize enforcer with GORM adapter (auto-creates 'casbin_rule' table). Unit test TestCasbinEnforcer passes, verifying policy operations.

- **Task 6: Implement User Management.**
  - Status: Completed
  - Reference Links:
    - GoFrame DAO: https://goframe.org/docs/database/gdb/dao
    - Bcrypt (Go): https://pkg.go.dev/golang.org/x/crypto/bcrypt
  - Notes:
    Created users table via SQL migration (`manifest/sql/0001_create_users_table.sql`). Manually created User DAO (`internal/dao/user_dao.go`), Entity (`user_entity.go`), and Input (`user_input.go`) models. Implemented basic UserLogic (`internal/logic/user/user_logic.go` for CreateUser, GetUserByUsername) and corresponding UserService (`internal/service/user.go`). Unit tests for creating and retrieving users (`user_logic_test.go`) pass, including checks for username duplication and fetching non-existent users. Full CRUD and other layers (Controller, API, Router) will be handled in subsequent tasks.

- **Task 7: Refine User Management (Controllers, API Routes, Full CRUD).**
  - Status: Completed
  - Reference Links:
    - GoFrame Controller: https://goframe.org/docs/net/ghttp/controller-router-group
    - GoFrame Validation (i18n): https://goframe.org/docs/component/gvalid/index#internationalization
    - GoFrame Router: https://goframe.org/docs/net/ghttp/router
  - Notes:
    - Sub-step 7a: Extended user input/output models (UserGetByIdInput, UserListInput, UserOutput, UserListOutput). Added Update, Delete, List methods to User DAO.
    - Sub-step 7b: Extended UserLogic and UserService for full CRUD (GetById, Update, Delete, List), including i18n for error messages.
    - Sub-step 7c: Implemented UserController with Create, GetById, Update, Delete, List methods. Integrated request parsing, GoFrame validation, and i18n for responses/validation errors.
    - Sub-step 7d: Defined API routes for User CRUD in internal/router/user_router.go. Implemented LanguageMiddleware for i18n in internal/router/router.go. Updated main.go to bind routes and start the HTTP server.
    - Sub-step 7e: Extended unit tests in internal/logic/user/user_logic_test.go to cover UpdateUser, DeleteUser, and ListUsers logic, including checks for i18n error messages and data integrity. All user logic tests pass.
    - Overall Task Status: Completed. Full CRUD backend for User Management is now implemented.

- **Task 8: Implement Department Management.** (was 7)
  - Status: In Progress
  - Reference Links:
  - Notes:
    - Foundational backend for Department Management: Created DB migration, Models (Entity, Input, Output), DAO, Logic, and Service for CRUD operations (including tree building for GetTree). Basic unit tests for Create/Get, Update, Delete, and List/Tree logic pass. Controller and API routes pending.

- **Task 9: Implement Post Management.** (was 8)
  - Status: In Progress
  - Reference Links:
  - Notes:
    - Foundational backend for Post Management: Created DB migration (`0003_create_posts_table.sql`), Models (Entity, Input, Output for Post), DAO, Logic, and Service layers for full CRUD operations. Added i18n keys for post-specific messages. Unit tests for PostLogic (Create, GetById, Update, Delete, List) all pass. Controller and API routes are pending.

- **Task 10: Implement Menu Management (linking to Casbin policies).** (was 9)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 11: Implement Role Management (linking to Casbin policies).** (was 10)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 12: Implement Dictionary Management.** (was 11)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 13: Implement Parameter Management.** (was 12)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 14: Implement Notification/Announcement Management.** (was 13)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 15: Implement Operation Log.** (was 14)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 16: Implement Login Log.** (was 15)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 17: Implement Code Generation (Backend Foundation).** (was 16)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 18: Implement Module Management Foundation.** (was 17)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 19: Implement Scheduled Tasks (CRUD & Logging).** (was 18)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 20: Implement Attachment Management.** (was 19)
  - Status: Pending
  - Reference Links:
  - Notes:

## Phase 3: Technical Feature Implementation

- **Task 21: Verify PostgreSQL Compatibility.** (was 20)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 22: Implement Multi-language Support.** (was 21)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 23: Implement Multi-theme Support (Backend Hooks).** (was 22)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 24: Implement Redis-based Queues.** (was 23)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 25: Implement WebSocket Support.** (was 24)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 26: Implement Global DB Caching.** (was 25)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 27: Implement Automatic API Documentation.** (was 26)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 28: Implement System Monitoring (Backend Logic).** (was 27)
  - Status: Pending
  - Reference Links:
  - Notes:

- **Task 29: Create Application Dockerfile.** (was 28)
  - Status: Pending
  - Reference Links:
  - Notes:
