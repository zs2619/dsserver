set SRC_DIR=../proto
set GO_DST_DIR=../pb
set BIN_Path_DIR=%cd%/protobuf/bin
set PATH=%PATH%;%BIN_Path_DIR%

protoc.exe --go_opt=paths=source_relative  --go-grpc_opt=paths=source_relative --proto_path=%SRC_DIR% --proto_path=%GRPC_SRC_DIR%  --go-grpc_out=%GO_DST_DIR% --go_out=%GO_DST_DIR%  %SRC_DIR%/*.proto