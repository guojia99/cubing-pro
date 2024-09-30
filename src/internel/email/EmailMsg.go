package email

type CodeTempData struct {
	Subject  string // 主题
	UserName string // 用户名
	BaseUrl  string // 域名

	// 操作邮箱
	Option         string // 操作
	OptionsTimeOut string // 超时时间
	OptionsCode    string // 验证码
	OptionsUrl     string // 连接

	// 通知
	Notify    string // 通知
	NotifyMsg string // 通知详情
	NotifyUrl string // 通知地址

	// 其他HTML
	HTML string // 其他HTML
}

const CodeTemp = `<!doctype html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <title></title>
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <style type="text/css">
        html, body, * {
            -webkit-text-size-adjust: none;
            text-size-adjust: none;
        }

        a {
            color: #1EB0F4;
            text-decoration: none;
        }

        a:hover {
            text-decoration: underline;
        }

        .CubingProLogo {
            font-size: 40px;
            border: none;
            border-radius: 20px;
            display: block;
            outline: none;
            text-decoration: none;
            width: 100%;
            height: 38px;
            color: #1EB0F4;
            text-align: center;
        }

        .CubingProCode {
            text-decoration: none;
            line-height: 100%;
            background: #2be5e5;
            color: white;
            font-family: Ubuntu, Helvetica, Arial, sans-serif;
            font-weight: normal;
            text-transform: none;
            width: 40%;
            text-align: center;
            height: 40px;
            font-size: 30px;
            margin-left: 30%;
            border-radius: 15px;
            display: flex;
            justify-content: center;
            align-items: center;
        }

        .CubingProUrlButton {
            width: 50% !important;
            margin-left: 25% !important;
            height: 60px !important;
            font-size: 25px !important;
            color: #F9F9F9 !important;
            text-decoration: none !important; /* 去除下划线 */
        }

        .CubingProMessages {
            cursor: auto;
            color: #737F8D;
            font-family: Helvetica Neue, Helvetica, Arial, Lucida Grande, sans-serif;
            font-size: 16px;
            line-height: 24px;
            text-align: left;
            width: 80%;
            margin-left: 10%;
            margin-top: 30px;
        }

        .CubingProFooter {
            cursor: auto;
            color: #747F8D;
            font-family: Helvetica Neue, Helvetica, Arial, Lucida Grande, sans-serif;
            font-size: 13px;
            line-height: 16px;
            text-align: left;
        }

        .CubingProHello {
            font-family: Helvetica Neue, Helvetica, Arial, Lucida Grande, sans-serif;
            font-weight: 500;
            font-size: 20px;
            color: #4F545C;
            letter-spacing: 0.27px;
        }

        .CubingProBase {
            vertical-align: top;
            display: inline-block;
            direction: ltr;
            font-size: 13px;
            text-align: left;
            width: 100%;
        }

        .CubingProBaseBody {
            max-width: 640px;
            margin: 0 auto;
            box-shadow: 0 1px 5px rgba(0, 0, 0, 0.1);
            border-radius: 4px;
            overflow: hidden;
            text-decoration: none; /* 去除下划线 */
        }
    </style>
</head>


<body style="background: #F9F9F9;">
<div style="margin-top: 30px; margin-bottom: 30px">
    <p class="CubingProLogo">
        CubingPro
    </p>
</div>


<div class="CubingProBaseBody">
    <div class="CubingProBase">
        <div class="CubingProMessages">
            <h1 style="text-align: center;">{{.Subject}}</h1>
            <h2 class="CubingProHello">
                你好! {{.UserName}}
            </h2>
            {{ if .Option}}
                {{/*                操作的验证码*/}}
                {{ if .OptionsCode }}
                    <p style="display: inline-block"> 你正在执行{{.Option}},以下是本次{{.Option}}的验证码,该验证码仅用于
                        <a href="{{.BaseUrl}}">CubingPro</a>的用户{{.Option}},
                        请妥善保管好该验证码,如果不是本人操作,请忽略该邮件,验证码将在: <a style="color: red">{{.OptionsTimeOut}}</a>到期</p>
                    <p class="CubingProCode" id="copyText">
                        {{.OptionsCode}}
                    </p>
                {{ end}}

                {{/*                操作链接*/}}
                {{ if .OptionsUrl }}
                    <p style="display: inline-block"> 你正在执行{{.Option}},以下是本次{{.Option}}的链接,该链接仅用于
                        <a href="{{.BaseUrl}}">CubingPro</a>的用户{{.Option}},
                        请点击下列按钮进行{{.Option}},如果不是本人操作,请忽略该邮件,链接将在: <a style="color: red">{{.OptionsTimeOut}}</a>到期</p>
                    <a href="{{.OptionsUrl}}" class="CubingProCode CubingProUrlButton">{{.Option}}</a>
                    <p style="color: grey; text-align: center">点击按钮查看详情</p>
                    <p>如果无法点击按钮, 请手动访问以下链接: <a href="{{.OptionsUrl}}">{{.OptionsUrl}}</a></p>
                {{ end }}
            {{ end}}


            {{ if .Notify }}
                <p style="display: inline-block">{{.NotifyMsg}}</p>
                {{ if .NotifyUrl }}
                    <p style="display: inline-block"> 你可以点击下列按钮查看详情</p>
                    <a href="{{.NotifyUrl}}" class="CubingProCode CubingProUrlButton">{{.Notify}}</a>
                    <p style="color: grey; text-align: center">点击按钮查看详情</p>
                    <p>如果无法点击按钮, 请手动访问以下链接: <a href="{{.NotifyUrl}}">{{.NotifyUrl}}</a></p>
                {{ end }}
            {{ end }}


            {{ if .HTML }}
                {{.HTML}}
            {{ end }}

            <div class="CubingProFooter">
                <p>本邮箱由后台自动发送，请勿回复此邮箱</p>
            </div>
        </div>
    </div>
</div>
</body>
`
