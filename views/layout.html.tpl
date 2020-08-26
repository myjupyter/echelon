<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<title>{{template "pagetitle" .}}</title>

	<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap.min.css" />
	<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/font-awesome/4.3.0/css/font-awesome.min.css" />

	<script type="text/javascript" src="https://code.jquery.com/jquery-2.1.3.min.js"></script>
	<script type="text/javascript" src="//maxcdn.bootstrapcdn.com/bootstrap/3.3.2/js/bootstrap.min.js"></script>
</head>
<body class="container-fluid" style="padding-top: 15px;">
	<nav class="navbar navbar-default">
		<div class="container-fluid">
			<div class="navbar-header">
				<a class="navbar-brand" href="/">Application</a>
			</div>

			<div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
                {{if .loggedin}}
                <ul class="nav navbar-nav navbar-left">
                    <li><a href="/foo">Foo</a></li>
                    <li><a href="/bar">Bar</a></li>
                    {{if .permission}}
                    <li><a href="/sigma">Sigma</a></li>
                    {{end}}
                </ul>
                {{end}}
				<ul class="nav navbar-nav navbar-right">
					{{if not .loggedin}}
					<li><a href="/auth/register">Register</a></li>
					<li><a href="/auth/recover">Recover</a></li>
					<li><a href="/auth/login"><i class="fa fa-sign-in"></i> Login</a></li>
					{{else}}
					<li class="dropdown">
						<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-expanded="false">Welcome {{.current_user_name}}! <span class="caret"></span></a>
						<ul class="dropdown-menu" role="menu">
							<li>
								<a href="/auth/logout">
									<i class="fa fa-sign-out"></i> Logout
								</a>
							</li>
						</ul>
					</li>
					{{end}}
				</ul>
			</div>
            </div>
            </nav>

    {{with .flash_success}}<div class="alert alert-success">{{.}}</div>{{end}}
    {{with .flash_error}}<div class="alert alert-danger">{{.}}</div>{{end}}
    {{template "yield" .}}
</body>
</html>
{{define "pagetitle"}}{{end}}
{{define "yield"}}{{end}}	

