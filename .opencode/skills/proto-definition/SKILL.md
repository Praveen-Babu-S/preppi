---
name: proto-definition
description: Use when creating, editing, or updating protobuf definitions (proto3) for the doubt-resolver platform. Trigger on requests like "add an RPC", "create a proto file", "update the question proto", "define messages for X". Enforces project proto conventions including package naming, pagination, timestamps, and error handling.
---

# Proto Definition

## Purpose

Create and maintain Protocol Buffers (proto3) definitions for the doubt-resolver microservices
following the project's proto conventions.

## When to Use

Use this skill when the user asks to:
- Create a new `.proto` file
- Add RPC methods to an existing service
- Add/modify message definitions
- Update field types or add pagination

## Proto File Location

- All protos live in `proto/<service>/v1/<service>.proto`
- One proto file per service in the `v1` folder

## Package Conventions

```proto
syntax = "proto3";

package <service>.v1;

option go_package = "github.com/<org>/doubt-resolver/proto/<service>/v1;v1";
```

Example:
```proto
syntax = "proto3";

package question.v1;

option go_package = "github.com/<org>/doubt-resolver/proto/question/v1;v1";
```

## Naming Conventions

- **Service name**: PascalCase, suffixed with `Service`: `QuestionService`
- **RPC methods**: PascalCase verbs: `CreateQuestion`, `GetQuestionById`
- **Request messages**: `<Method>Request`: `CreateQuestionRequest`
- **Response messages**: `<Method>Response`: `CreateQuestionResponse`
- **Enums**: PascalCase type name, `UPPER_SNAKE_CASE` values, always `0 = <TYPE>_UNSPECIFIED`
- **Fields**: `snake_case`
- **Package**: `lowercase` single word

## Message Conventions

### Timestamps
Always use `google.protobuf.Timestamp`, never strings:
```proto
import "google.protobuf/timestamp.proto";

message Question {
  google.protobuf.Timestamp created_at = 3;
}
```

### Partial Updates
Use `google.protobuf.FieldMask` for partial updates:
```proto
import "google.protobuf/field_mask.proto";

message UpdateQuestionRequest {
  string question_id = 1;
  google.protobuf.FieldMask update_mask = 2;
}
```

### Pagination
List requests use `page_token` + `page_size`; responses use `next_page_token`:
```proto
message ListQuestionsRequest {
  string student_id = 1;
  string page_token = 2;   // opaque cursor
  int32 page_size = 3;     // 1-100, default 20
}

message ListQuestionsResponse {
  repeated Question questions = 1;
  string next_page_token = 2;  // empty if no more
}
```

### Field Number Rules
- 1-15: frequently used fields (lower wire overhead, 1 byte)
- 16+: less frequent fields
- Never reuse a field number; never change a field type (forward/backward compat)

### Enums
```proto
enum QuestionStatus {
  QUESTION_STATUS_UNSPECIFIED = 0;
  QUESTION_STATUS_OPEN = 1;
  QUESTION_STATUS_ASSIGNED = 2;
  QUESTION_STATUS_IN_PROGRESS = 3;
  QUESTION_STATUS_ANSWERED = 4;
  QUESTION_STATUS_ESCALATED = 5;
}
```

## RPC Conventions

```proto
service QuestionService {
  // Create a new question (student only)
  rpc CreateQuestion(CreateQuestionRequest) returns (CreateQuestionResponse);
  // Get a question by ID with its solution and comments
  rpc GetQuestionById(GetQuestionByIdRequest) returns (GetQuestionByIdResponse);
  // List questions for a student with pagination
  rpc ListQuestions(ListQuestionsRequest) returns (ListQuestionsResponse);
  // Streaming: real-time follow-up updates on a question
  rpc StreamQuestion(StreamQuestionRequest) returns (stream QuestionUpdate);
}
```

- All RPCs return **typed response messages**, never bare errors as return types
- Support both unary and streaming RPCs where real-time is needed
- Include auth in metadata (grpc metadata), not in request messages, unless it's a resource owner reference

## Template

```proto
syntax = "proto3";

package <service>.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/field_mask.proto";

option go_package = "github.com/<org>/doubt-resolver/proto/<service>/v1;v1";

service <Name>Service {
  rpc Create( CreateRequest) returns ( CreateResponse);
}

message Entity {
  string id = 1;
  string name = 2;
  EntityStatus status = 3;
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}

enum EntityStatus {
  ENTITY_STATUS_UNSPECIFIED = 0;
  ENTITY_STATUS_ACTIVE = 1;
  ENTITY_STATUS_INACTIVE = 2;
}

message CreateRequest {
  string name = 1;
}

message CreateResponse {
  Entity entity = 1;
}
```

## After Writing

After creating or editing proto files, remind the user to regenerate Go code:
```
make proto-gen
```
Or run the protoc command directly for the modified proto.
