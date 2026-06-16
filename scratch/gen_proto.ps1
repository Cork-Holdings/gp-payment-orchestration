# Code generation script to compile Protobuf definitions using source_relative paths

$protoc = "c:\Users\virus\Desktop\geepay\bin\protoc.exe"
$protoDir = "c:\Users\virus\Desktop\geepay\gp-payment-orchestration\proto"

$authDir = "c:\Users\virus\Desktop\geepay\gp_auth"
$orchDir = "c:\Users\virus\Desktop\geepay\gp-payment-orchestration"

# Clean any incorrectly generated github.com subdirectories
if (Test-Path "$authDir\github.com") { Remove-Item -Recurse -Force "$authDir\github.com" }
if (Test-Path "$orchDir\github.com") { Remove-Item -Recurse -Force "$orchDir\github.com" }

# Create output directories for gp_auth
New-Item -ItemType Directory -Force -Path "$authDir\proto\authpb"

# Create output directories for gp-payment-orchestration
New-Item -ItemType Directory -Force -Path "$orchDir\proto\authpb"
New-Item -ItemType Directory -Force -Path "$orchDir\proto\ledgerpb"
New-Item -ItemType Directory -Force -Path "$orchDir\proto\settlementpb"
New-Item -ItemType Directory -Force -Path "$orchDir\proto\floatpb"
New-Item -ItemType Directory -Force -Path "$orchDir\proto\kycpb"

# 1. Compile for gp_auth
Write-Host "Compiling auth.proto for gp_auth..."
& $protoc --proto_path=$protoDir --go_out="$authDir\proto\authpb" --go_opt=paths=source_relative --go-grpc_out="$authDir\proto\authpb" --go-grpc_opt=paths=source_relative "$protoDir\auth.proto"

# 2. Compile for gp-payment-orchestration
Write-Host "Compiling protos for gp-payment-orchestration..."
& $protoc --proto_path=$protoDir --go_out="$orchDir\proto\authpb" --go_opt=paths=source_relative --go-grpc_out="$orchDir\proto\authpb" --go-grpc_opt=paths=source_relative "$protoDir\auth.proto"
& $protoc --proto_path=$protoDir --go_out="$orchDir\proto\ledgerpb" --go_opt=paths=source_relative --go-grpc_out="$orchDir\proto\ledgerpb" --go-grpc_opt=paths=source_relative "$protoDir\ledger.proto"
& $protoc --proto_path=$protoDir --go_out="$orchDir\proto\settlementpb" --go_opt=paths=source_relative --go-grpc_out="$orchDir\proto\settlementpb" --go-grpc_opt=paths=source_relative "$protoDir\settlement.proto"
& $protoc --proto_path=$protoDir --go_out="$orchDir\proto\floatpb" --go_opt=paths=source_relative --go-grpc_out="$orchDir\proto\floatpb" --go-grpc_opt=paths=source_relative "$protoDir\float.proto"
& $protoc --proto_path=$protoDir --go_out="$orchDir\proto\kycpb" --go_opt=paths=source_relative --go-grpc_out="$orchDir\proto\kycpb" --go-grpc_opt=paths=source_relative "$protoDir\kyc.proto"

Write-Host "Protobuf code generation completed successfully!"
