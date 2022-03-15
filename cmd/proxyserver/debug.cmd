cd web
call pnpm run build
copy .\dist\inject.min.js ..\..\..\pkg\server\inject.min.js
cd ..
go run main.go