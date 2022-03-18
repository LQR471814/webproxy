cd web
call pnpm run build
copy .\dist\post.min.js ..\..\..\pkg\server\post.min.js
copy .\dist\pre.min.js ..\..\..\pkg\server\pre.min.js
cd ..
go run main.go