param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [string]$ConfigPath = "config/config.yaml",
    [string]$MysqlExe = "mysql",
    [switch]$SkipDbMutations
)

$ErrorActionPreference = "Stop"

$script:PatchDir = $PSScriptRoot
$script:BackendRoot = (Resolve-Path (Join-Path $script:PatchDir "..\..\..")).Path
if (![System.IO.Path]::IsPathRooted($ConfigPath)) {
    $ConfigPath = Join-Path $script:BackendRoot $ConfigPath
}
$ConfigPath = (Resolve-Path $ConfigPath).Path

$script:Total = 0
$script:Passed = 0
$script:Failed = 0
$script:Skipped = 0

function Write-CaseResult {
    param(
        [string]$Name,
        [string]$Status,
        [string]$Detail = ""
    )

    $script:Total++
    switch ($Status) {
        "PASS" { $script:Passed++; Write-Host "[PASS] $Name" -ForegroundColor Green }
        "FAIL" { $script:Failed++; Write-Host "[FAIL] $Name $Detail" -ForegroundColor Red }
        "SKIP" { $script:Skipped++; Write-Host "[SKIP] $Name $Detail" -ForegroundColor Yellow }
    }
}

function ConvertTo-BodyJson {
    param([object]$Body)

    if ($null -eq $Body) {
        return $null
    }
    return ($Body | ConvertTo-Json -Depth 10 -Compress)
}

function Invoke-Api {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Path,
        [object]$Body = $null,
        [string]$Token = "",
        [int]$ExpectedCode = 0,
        [int]$ExpectedHttpStatus = 200
    )

    $headers = @{}
    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    $json = ConvertTo-BodyJson $Body
    $statusCode = $null
    $content = $null

    try {
        if ($null -eq $json) {
            $res = Invoke-WebRequest -UseBasicParsing -Method $Method -Uri "$BaseUrl$Path" -Headers $headers
        } else {
            $res = Invoke-WebRequest -UseBasicParsing -Method $Method -Uri "$BaseUrl$Path" -Headers $headers -ContentType "application/json" -Body $json
        }
        $statusCode = [int]$res.StatusCode
        $content = $res.Content
    } catch {
        if ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
            if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
                $content = $_.ErrorDetails.Message
            } else {
                $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
                $content = $reader.ReadToEnd()
                $reader.Close()
            }
        } else {
            Write-CaseResult $Name "FAIL" $_.Exception.Message
            return $null
        }
    }

    try {
        $bodyObj = $content | ConvertFrom-Json
    } catch {
        Write-CaseResult $Name "FAIL" "response is not JSON: $content"
        return $null
    }

    if ($statusCode -ne $ExpectedHttpStatus -or [int]$bodyObj.code -ne $ExpectedCode) {
        Write-CaseResult $Name "FAIL" "expected http/code $ExpectedHttpStatus/$ExpectedCode, got $statusCode/$($bodyObj.code), message=$($bodyObj.message)"
        return $bodyObj
    }

    Write-CaseResult $Name "PASS"
    return $bodyObj
}

function Get-YamlValue {
    param(
        [string[]]$Lines,
        [string]$Section,
        [string]$Key,
        [string]$Default = ""
    )

    $inSection = $false
    foreach ($line in $Lines) {
        if ($line -match "^\s*$Section\s*:\s*$") {
            $inSection = $true
            continue
        }
        if ($inSection -and $line -match "^\S") {
            $inSection = $false
        }
        if ($inSection -and $line -match "^\s+$Key\s*:\s*(.+?)\s*$") {
            return $Matches[1].Trim(" `"`'")
        }
    }
    return $Default
}

function Load-DbConfig {
    if (!(Test-Path $ConfigPath)) {
        return $null
    }

    $lines = Get-Content -Encoding UTF8 $ConfigPath
    return @{
        host = Get-YamlValue $lines "database" "host" "127.0.0.1"
        port = Get-YamlValue $lines "database" "port" "3306"
        user = Get-YamlValue $lines "database" "user" "root"
        password = Get-YamlValue $lines "database" "password" ""
        name = Get-YamlValue $lines "database" "name" "cau_used_goods"
    }
}

function Test-CommandAvailable {
    param([string]$Command)
    return $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
}

$script:DbConfig = Load-DbConfig
$script:CanUseMysql = $false
if (!$SkipDbMutations -and $script:DbConfig -and ((Test-CommandAvailable $MysqlExe) -or (Test-CommandAvailable "go"))) {
    $script:CanUseMysql = $true
}

function Invoke-GoDbExec {
    param([string]$Sql)

    if (!(Test-CommandAvailable "go")) {
        throw "go command is unavailable"
    }

    $dbExecSource = @'
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"cau-used-goods-app/backend/internal/config"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "config file path")
	sqlText := flag.String("sql", "", "SQL statement to execute")
	flag.Parse()

	if *sqlText == "" {
		fmt.Fprintln(os.Stderr, "sql is required")
		os.Exit(2)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "open mysql: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if _, err := db.Exec(*sqlText); err != nil {
		fmt.Fprintf(os.Stderr, "exec sql: %v\n", err)
		os.Exit(1)
	}
}
'@
    $tempDbExecPath = Join-Path $script:BackendRoot "test-dbexec-$PID.go"
    Set-Content -Encoding UTF8 -Path $tempDbExecPath -Value $dbExecSource

    $oldErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        Push-Location $script:BackendRoot
        $output = & go run $tempDbExecPath -config $ConfigPath -sql $Sql 2>&1
    } finally {
        Pop-Location
        Remove-Item -Force -ErrorAction SilentlyContinue $tempDbExecPath
        $ErrorActionPreference = $oldErrorActionPreference
    }
    if ($LASTEXITCODE -ne 0) {
        throw "go dbexec failed: $output"
    }
    return $output
}

function Invoke-MySql {
    param([string]$Sql)

    if (!$script:CanUseMysql) {
        throw "database mutation is disabled"
    }

    if (!(Test-CommandAvailable $MysqlExe)) {
        return Invoke-GoDbExec $Sql
    }

    $args = @(
        "-h", $script:DbConfig.host,
        "-P", $script:DbConfig.port,
        "-u", $script:DbConfig.user,
        "--ssl-mode=DISABLED",
        "--default-character-set=utf8mb4",
        "-D", $script:DbConfig.name,
        "-N",
        "-e", $Sql
    )
    if ($script:DbConfig.password) {
        $args = @("-p$($script:DbConfig.password)") + $args
    }

    $oldErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $output = & $MysqlExe @args 2>&1
    } finally {
        $ErrorActionPreference = $oldErrorActionPreference
    }
    if ($LASTEXITCODE -ne 0) {
        return Invoke-GoDbExec $Sql
    }
    return @($output | Where-Object { "$_" -notlike "mysql: [Warning]*" })
}

function Escape-Sql {
    param([string]$Value)
    return $Value.Replace("\", "\\").Replace("'", "''")
}

function Set-TestUserState {
    param(
        [string]$OpenID,
        [string]$Role = "",
        [string]$AuthStatus = "",
        [string]$AccountStatus = "",
        [string]$StudentID = $null,
        [switch]$ClearStudentInfo
    )

    if (!$script:CanUseMysql) {
        return
    }

    $sets = @("update_time = NOW()")
    if ($Role) { $sets += "role = '$(Escape-Sql $Role)'" }
    if ($AuthStatus) { $sets += "auth_status = '$(Escape-Sql $AuthStatus)'" }
    if ($AccountStatus) { $sets += "account_status = '$(Escape-Sql $AccountStatus)'" }
    if ($ClearStudentInfo) {
        $sets += "student_id = NULL"
        $sets += "real_name = NULL"
        $sets += "college = NULL"
    } elseif (![string]::IsNullOrEmpty($StudentID)) {
        $sets += "student_id = '$(Escape-Sql $StudentID)'"
    }

    $sql = "UPDATE users SET $($sets -join ', ') WHERE openid = '$(Escape-Sql $OpenID)' AND is_deleted = 0;"
    Invoke-MySql $sql | Out-Null
}

function New-DevLogin {
    param(
        [string]$OpenID,
        [string]$Role = "USER",
        [string]$Name = "dev-login"
    )

    $res = Invoke-Api -Name $Name -Method "POST" -Path "/auth/dev-login" -Body @{ openid = $OpenID; role = $Role } -ExpectedCode 0 -ExpectedHttpStatus 200
    if ($res -and $res.data) {
        return @{
            token = [string]$res.data.token
            userId = [uint64]$res.data.user.id
            user = $res.data.user
        }
    }
    return $null
}

function Submit-StudentVerify {
    param(
        [string]$Name,
        [string]$Token,
        [string]$StudentID,
        [int]$ExpectedCode = 0,
        [int]$ExpectedHttpStatus = 200
    )

    return Invoke-Api -Name $Name -Method "POST" -Path "/users/student-verify" -Token $Token -Body @{
        studentId = $StudentID
        realName = "测试学生"
        college = "信息学院"
    } -ExpectedCode $ExpectedCode -ExpectedHttpStatus $ExpectedHttpStatus
}

function Invoke-CurlUpload {
    param(
        [string]$Name,
        [string]$Token,
        [string]$FilePath,
        [int]$ExpectedCode,
        [int]$ExpectedHttpStatus
    )

    if (!(Test-CommandAvailable "curl.exe")) {
        Write-CaseResult $Name "SKIP" "curl.exe not found"
        return
    }

    $responseFile = Join-Path $env:TEMP "cau_api_test_response_$([guid]::NewGuid().ToString('N')).json"
    $status = & curl.exe -s -o $responseFile -w "%{http_code}" `
        -X POST "$BaseUrl/users/avatar" `
        -H "Authorization: Bearer $Token" `
        -F "avatar=@$FilePath"

    $content = Get-Content -Raw -Encoding UTF8 $responseFile
    Remove-Item $responseFile -Force -ErrorAction SilentlyContinue

    try {
        $obj = $content | ConvertFrom-Json
    } catch {
        Write-CaseResult $Name "FAIL" "response is not JSON: $content"
        return
    }

    if ([int]$status -ne $ExpectedHttpStatus -or [int]$obj.code -ne $ExpectedCode) {
        Write-CaseResult $Name "FAIL" "expected http/code $ExpectedHttpStatus/$ExpectedCode, got $status/$($obj.code), message=$($obj.message)"
        return
    }
    Write-CaseResult $Name "PASS"
}

Write-Host "Login/auth API test started. BaseUrl=$BaseUrl"
if (!$script:CanUseMysql) {
    Write-Host "DB mutation tests will be skipped. Provide mysql.exe and config/config.yaml, or run without -SkipDbMutations." -ForegroundColor Yellow
}

$stamp = Get-Date -Format "yyyyMMddHHmmss"

$user = New-DevLogin -OpenID "api_test_user_$stamp" -Role "USER" -Name "dev-login 普通用户"
$admin = New-DevLogin -OpenID "api_test_admin_$stamp" -Role "ADMIN" -Name "dev-login 管理员"

Invoke-Api -Name "dev-login 非法 role" -Method "POST" -Path "/auth/dev-login" -Body @{ openid = "api_test_bad_role_$stamp"; role = "ROOT" } -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
Invoke-Api -Name "dev-login openid 过长" -Method "POST" -Path "/auth/dev-login" -Body @{ openid = ("a" * 65); role = "USER" } -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
Invoke-Api -Name "wechat-login 缺少 code" -Method "POST" -Path "/auth/wechat-login" -Body @{} -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
Invoke-Api -Name "wechat-login code 过长" -Method "POST" -Path "/auth/wechat-login" -Body @{ code = ("x" * 129) } -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
Invoke-Api -Name "无 token 查自己" -Method "GET" -Path "/users/me" -ExpectedCode 401 -ExpectedHttpStatus 401 | Out-Null
Invoke-Api -Name "错误 token 查自己" -Method "GET" -Path "/users/me" -Token "bad.token.value" -ExpectedCode 401 -ExpectedHttpStatus 401 | Out-Null

if ($user) {
    Invoke-Api -Name "普通用户访问管理员接口" -Method "GET" -Path "/admin/users/student-verifications" -Token $user.token -ExpectedCode 403 -ExpectedHttpStatus 403 | Out-Null
    Invoke-Api -Name "修改昵称成功" -Method "PUT" -Path "/users/profile" -Token $user.token -Body @{ nickname = "测试用户"; phone = "13800000000" } -ExpectedCode 0 -ExpectedHttpStatus 200 | Out-Null
    Invoke-Api -Name "昵称为空" -Method "PUT" -Path "/users/profile" -Token $user.token -Body @{ nickname = "" } -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
    Invoke-Api -Name "手机号格式错误" -Method "PUT" -Path "/users/profile" -Token $user.token -Body @{ phone = "123" } -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
}

if ($script:CanUseMysql) {
    $demotedAdmin = New-DevLogin -OpenID "api_test_demoted_admin_$stamp" -Role "ADMIN" -Name "准备被降权管理员"
    if ($demotedAdmin) {
        Set-TestUserState -OpenID "api_test_demoted_admin_$stamp" -Role "USER" -AccountStatus "NORMAL"
        Invoke-Api -Name "被降权管理员访问管理员接口" -Method "GET" -Path "/admin/users/student-verifications" -Token $demotedAdmin.token -ExpectedCode 403 -ExpectedHttpStatus 403 | Out-Null
    }

    $disabledAdmin = New-DevLogin -OpenID "api_test_disabled_admin_$stamp" -Role "ADMIN" -Name "准备异常管理员"
    if ($disabledAdmin) {
        Set-TestUserState -OpenID "api_test_disabled_admin_$stamp" -Role "ADMIN" -AccountStatus "DISABLED"
        Invoke-Api -Name "管理员账号异常访问后台" -Method "GET" -Path "/admin/users/student-verifications" -Token $disabledAdmin.token -ExpectedCode 403 -ExpectedHttpStatus 403 | Out-Null
    }

    $disabledUser = New-DevLogin -OpenID "api_test_disabled_user_$stamp" -Role "USER" -Name "准备禁用用户"
    if ($disabledUser) {
        Set-TestUserState -OpenID "api_test_disabled_user_$stamp" -AccountStatus "DISABLED"
        Invoke-Api -Name "禁用用户修改资料" -Method "PUT" -Path "/users/profile" -Token $disabledUser.token -Body @{ nickname = "disabled" } -ExpectedCode 403 -ExpectedHttpStatus 403 | Out-Null
    }
} else {
    Write-CaseResult "被降权管理员访问管理员接口" "SKIP" "requires DB mutation"
    Write-CaseResult "管理员账号异常访问后台" "SKIP" "requires DB mutation"
    Write-CaseResult "禁用用户修改资料" "SKIP" "requires DB mutation"
}

if ($user) {
    $tmpDir = Join-Path $env:TEMP "cau_api_test_$stamp"
    New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
    $bigAvatar = Join-Path $tmpDir "big.png"
    $badAvatar = Join-Path $tmpDir "bad.txt"
    $tinyAvatar = Join-Path $tmpDir "tiny.png"
    [System.IO.File]::WriteAllBytes($bigAvatar, (New-Object byte[] (2 * 1024 * 1024 + 1)))
    Set-Content -Encoding UTF8 -Path $badAvatar -Value "not image"
    Set-Content -Encoding UTF8 -Path $tinyAvatar -Value "png"

    Invoke-CurlUpload -Name "上传超大头像" -Token $user.token -FilePath $bigAvatar -ExpectedCode 400 -ExpectedHttpStatus 400
    Invoke-CurlUpload -Name "上传非图片头像" -Token $user.token -FilePath $badAvatar -ExpectedCode 400 -ExpectedHttpStatus 400

    if ($script:CanUseMysql) {
        $disabledAvatarUser = New-DevLogin -OpenID "api_test_disabled_avatar_$stamp" -Role "USER" -Name "准备禁用头像用户"
        if ($disabledAvatarUser) {
            Set-TestUserState -OpenID "api_test_disabled_avatar_$stamp" -AccountStatus "DISABLED"
            Invoke-CurlUpload -Name "禁用用户上传头像" -Token $disabledAvatarUser.token -FilePath $tinyAvatar -ExpectedCode 403 -ExpectedHttpStatus 403
        }
    } else {
        Write-CaseResult "禁用用户上传头像" "SKIP" "requires DB mutation"
    }
}

if ($script:CanUseMysql) {
    $verifyUser = New-DevLogin -OpenID "api_test_verify_$stamp" -Role "USER" -Name "准备认证用户"
    if ($verifyUser) {
        Set-TestUserState -OpenID "api_test_verify_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
        Submit-StudentVerify -Name "提交学生认证成功" -Token $verifyUser.token -StudentID "T$($stamp.Substring(4,10))" | Out-Null
    }

    $badStudentUser = New-DevLogin -OpenID "api_test_bad_student_$stamp" -Role "USER" -Name "准备学号格式用户"
    if ($badStudentUser) {
        Set-TestUserState -OpenID "api_test_bad_student_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
        Submit-StudentVerify -Name "学号格式错误" -Token $badStudentUser.token -StudentID "bad!" -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
    }

    $duplicateID = "D$($stamp.Substring(4,10))"
    $dupUser1 = New-DevLogin -OpenID "api_test_dup1_$stamp" -Role "USER" -Name "准备重复学号用户1"
    $dupUser2 = New-DevLogin -OpenID "api_test_dup2_$stamp" -Role "USER" -Name "准备重复学号用户2"
    if ($dupUser1 -and $dupUser2) {
        Set-TestUserState -OpenID "api_test_dup1_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
        Set-TestUserState -OpenID "api_test_dup2_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
        Submit-StudentVerify -Name "准备重复学号占用" -Token $dupUser1.token -StudentID $duplicateID | Out-Null
        Submit-StudentVerify -Name "学号重复" -Token $dupUser2.token -StudentID $duplicateID -ExpectedCode 409 -ExpectedHttpStatus 409 | Out-Null
    }

    $pendingUser = New-DevLogin -OpenID "api_test_pending_$stamp" -Role "USER" -Name "准备 PENDING 用户"
    if ($pendingUser) {
        Set-TestUserState -OpenID "api_test_pending_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
        Submit-StudentVerify -Name "准备 PENDING 状态" -Token $pendingUser.token -StudentID "P$($stamp.Substring(4,10))" | Out-Null
        Submit-StudentVerify -Name "PENDING 重复提交" -Token $pendingUser.token -StudentID "P2$($stamp.Substring(5,9))" -ExpectedCode 409 -ExpectedHttpStatus 409 | Out-Null
    }

    $verifiedUser = New-DevLogin -OpenID "api_test_verified_$stamp" -Role "USER" -Name "准备 VERIFIED 用户"
    if ($verifiedUser) {
        Set-TestUserState -OpenID "api_test_verified_$stamp" -AuthStatus "VERIFIED" -AccountStatus "NORMAL" -StudentID "V$($stamp.Substring(4,10))"
        Submit-StudentVerify -Name "VERIFIED 重复提交" -Token $verifiedUser.token -StudentID "V2$($stamp.Substring(5,9))" -ExpectedCode 409 -ExpectedHttpStatus 409 | Out-Null
    }

    $rejectedUser = New-DevLogin -OpenID "api_test_rejected_$stamp" -Role "USER" -Name "准备 REJECTED 用户"
    if ($rejectedUser) {
        Set-TestUserState -OpenID "api_test_rejected_$stamp" -AuthStatus "REJECTED" -AccountStatus "NORMAL" -ClearStudentInfo
        Submit-StudentVerify -Name "REJECTED 重新提交" -Token $rejectedUser.token -StudentID "R$($stamp.Substring(4,10))" | Out-Null
    }

    if ($admin) {
        $approveUser = New-DevLogin -OpenID "api_test_approve_$stamp" -Role "USER" -Name "准备审核通过用户"
        if ($approveUser) {
            Set-TestUserState -OpenID "api_test_approve_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
            Submit-StudentVerify -Name "准备审核通过 PENDING" -Token $approveUser.token -StudentID "A$($stamp.Substring(4,10))" | Out-Null
            Invoke-Api -Name "管理员审核通过" -Method "PUT" -Path "/admin/users/$($approveUser.userId)/student-verify" -Token $admin.token -Body @{ authStatus = "VERIFIED"; description = "学生认证审核通过" } -ExpectedCode 0 -ExpectedHttpStatus 200 | Out-Null
        }

        $rejectUser = New-DevLogin -OpenID "api_test_reject_$stamp" -Role "USER" -Name "准备审核驳回用户"
        if ($rejectUser) {
            Set-TestUserState -OpenID "api_test_reject_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
            Submit-StudentVerify -Name "准备审核驳回 PENDING" -Token $rejectUser.token -StudentID "J$($stamp.Substring(4,10))" | Out-Null
            Invoke-Api -Name "管理员审核驳回" -Method "PUT" -Path "/admin/users/$($rejectUser.userId)/student-verify" -Token $admin.token -Body @{ authStatus = "REJECTED"; description = "学号或姓名信息不匹配" } -ExpectedCode 0 -ExpectedHttpStatus 200 | Out-Null
        }

        $rejectNoDescUser = New-DevLogin -OpenID "api_test_reject_empty_$stamp" -Role "USER" -Name "准备驳回无说明用户"
        if ($rejectNoDescUser) {
            Set-TestUserState -OpenID "api_test_reject_empty_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
            Submit-StudentVerify -Name "准备驳回无说明 PENDING" -Token $rejectNoDescUser.token -StudentID "E$($stamp.Substring(4,10))" | Out-Null
            Invoke-Api -Name "驳回无说明" -Method "PUT" -Path "/admin/users/$($rejectNoDescUser.userId)/student-verify" -Token $admin.token -Body @{ authStatus = "REJECTED"; description = "" } -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
        }

        $longDescUser = New-DevLogin -OpenID "api_test_long_desc_$stamp" -Role "USER" -Name "准备说明过长用户"
        if ($longDescUser) {
            Set-TestUserState -OpenID "api_test_long_desc_$stamp" -AuthStatus "UNVERIFIED" -AccountStatus "NORMAL" -ClearStudentInfo
            Submit-StudentVerify -Name "准备说明过长 PENDING" -Token $longDescUser.token -StudentID "L$($stamp.Substring(4,10))" | Out-Null
            Invoke-Api -Name "审核说明过长" -Method "PUT" -Path "/admin/users/$($longDescUser.userId)/student-verify" -Token $admin.token -Body @{ authStatus = "REJECTED"; description = ("x" * 501) } -ExpectedCode 400 -ExpectedHttpStatus 400 | Out-Null
        }

        $nonPendingUser = New-DevLogin -OpenID "api_test_non_pending_$stamp" -Role "USER" -Name "准备非 PENDING 用户"
        if ($nonPendingUser) {
            Set-TestUserState -OpenID "api_test_non_pending_$stamp" -AuthStatus "VERIFIED" -AccountStatus "NORMAL" -StudentID "N$($stamp.Substring(4,10))"
            Invoke-Api -Name "审核非 PENDING 用户" -Method "PUT" -Path "/admin/users/$($nonPendingUser.userId)/student-verify" -Token $admin.token -Body @{ authStatus = "VERIFIED"; description = "重复审核" } -ExpectedCode 409 -ExpectedHttpStatus 409 | Out-Null
        }

        Invoke-Api -Name "管理员审核自己" -Method "PUT" -Path "/admin/users/$($admin.userId)/student-verify" -Token $admin.token -Body @{ authStatus = "VERIFIED"; description = "self" } -ExpectedCode 403 -ExpectedHttpStatus 403 | Out-Null
    }
} else {
    Write-CaseResult "提交学生认证成功" "SKIP" "requires DB mutation"
    Write-CaseResult "学号格式错误" "SKIP" "requires DB mutation"
    Write-CaseResult "学号重复" "SKIP" "requires DB mutation"
    Write-CaseResult "PENDING 重复提交" "SKIP" "requires DB mutation"
    Write-CaseResult "VERIFIED 重复提交" "SKIP" "requires DB mutation"
    Write-CaseResult "REJECTED 重新提交" "SKIP" "requires DB mutation"
    Write-CaseResult "管理员审核通过" "SKIP" "requires DB mutation"
    Write-CaseResult "管理员审核驳回" "SKIP" "requires DB mutation"
    Write-CaseResult "驳回无说明" "SKIP" "requires DB mutation"
    Write-CaseResult "审核说明过长" "SKIP" "requires DB mutation"
    Write-CaseResult "审核非 PENDING 用户" "SKIP" "requires DB mutation"
    Write-CaseResult "管理员审核自己" "SKIP" "requires DB mutation"
}

Write-Host ""
Write-Host "Summary: total=$script:Total passed=$script:Passed failed=$script:Failed skipped=$script:Skipped"
if ($script:Failed -gt 0) {
    exit 1
}
exit 0
