echo on

::指定起始文件夹
set DIR="%cd%"

go get all
go mod tidy

cd %DIR%
go get all
go mod tidy

cd %DIR%/aggregator
go get all
go mod tidy

cd %DIR%/bank_card
go get all
go mod tidy

cd %DIR%/code_generator
go get all
go mod tidy

cd %DIR%/copierutil
go get all
go mod tidy

cd %DIR%/ddl_parser
go get all
go mod tidy

cd %DIR%/eventloop
go get all
go mod tidy

cd %DIR%/geoip
go get all
go mod tidy

cd %DIR%/id
go get all
go mod tidy

cd %DIR%/jwtutil
go get all
go mod tidy

cd %DIR%/mapper
go get all
go mod tidy

cd %DIR%/name_generator
go get all
go mod tidy

cd %DIR%/password
go get all
go mod tidy

cd %DIR%/query_parser
go get all
go mod tidy

cd %DIR%/slug
go get all
go mod tidy

cd %DIR%/translator
go get all
go mod tidy

cd %DIR%/translator/alibaba
go get all
go mod tidy

cd %DIR%/translator/baidu
go get all
go mod tidy

cd %DIR%/translator/google
go get all
go mod tidy

cd %DIR%/translator/volc
go get all
go mod tidy