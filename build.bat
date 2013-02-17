go install github.com/AlekSi/nut
@if ERRORLEVEL 1 exit /B %ERRORLEVEL%

go build -o gonut.exe github.com/AlekSi/nut/nut
@if ERRORLEVEL 1 exit /B %ERRORLEVEL%

@pushd ..\test_nut1
..\nut\gonut.exe pack
@if ERRORLEVEL 1 exit /B %ERRORLEVEL%
@popd

go test -v github.com/AlekSi/nut -gocheck.v
@if ERRORLEVEL 1 exit /B %ERRORLEVEL%

go test -v github.com/AlekSi/nut/nut -gocheck.v
@if ERRORLEVEL 1 exit /B %ERRORLEVEL%

go test -v -short github.com/AlekSi/nut/integration_test -gocheck.v
@if ERRORLEVEL 1 exit /B %ERRORLEVEL%
