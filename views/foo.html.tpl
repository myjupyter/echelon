{{define "pagetitle"}}Foo{{end}}

{{$loggedin := .loggedin}}
{{if $loggedin}}
    Logged
{{end}}

Foo text


