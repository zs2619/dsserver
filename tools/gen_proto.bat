
set SRC_DIR=../proto
set GRPC_SRC_DIR=../proto/server
set PB_DST_DIR=../Server/pb/msg.pb
set GO_DST_DIR=../Server/pb
set BIN_Paht_DIR=%cd%/protobuf/bin
set PATH=%PATH%;%BIN_Paht_DIR%


@REM protoc.exe --proto_path=%SRC_DIR%  --go_out=%GO_DST_DIR% --cpp_out=%CPP_DST_DIR% %SRC_DIR%/*.proto
protoc.exe --proto_path=%SRC_DIR%  --descriptor_set_out=%PB_DST_DIR% --go_out=%GO_DST_DIR%  %SRC_DIR%/*.proto

protoc.exe --proto_path=%GRPC_SRC_DIR%    --go-grpc_out=%GO_DST_DIR% --go_out=%GO_DST_DIR%  %GRPC_SRC_DIR%/*.proto

