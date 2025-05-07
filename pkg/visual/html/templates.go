package html

// htmlTemplate is the main HTML template for package documentation
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - {{.Package.Name}}</title>
    {{if .IncludeCSS}}
    <style>
        /* Base styles */
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 20px;
            background-color: #f8f9fa;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 20px;
            border-radius: 5px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
        }
        
        h1, h2, h3, h4 {
            margin-top: 1.5em;
            margin-bottom: 0.5em;
            font-weight: 600;
            color: #2c3e50;
        }
        
        h1 {
            font-size: 2em;
            padding-bottom: 0.3em;
            border-bottom: 1px solid #eaecef;
        }
        
        h2 {
            font-size: 1.5em;
            padding-bottom: 0.3em;
            border-bottom: 1px solid #eaecef;
        }
        
        h3 {
            font-size: 1.25em;
        }
        
        a {
            color: #0366d6;
            text-decoration: none;
        }
        
        a:hover {
            text-decoration: underline;
        }
        
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 1em 0;
        }
        
        th, td {
            padding: 8px 12px;
            text-align: left;
            border-bottom: 1px solid #e1e4e8;
        }
        
        th {
            background-color: #f6f8fa;
            font-weight: 600;
        }
        
        tr:hover {
            background-color: #f6f8fa;
        }
        
        .code {
            font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
            background-color: #f6f8fa;
            padding: 16px;
            border-radius: 3px;
            overflow-x: auto;
            font-size: 13px;
            line-height: 1.45;
            white-space: pre-wrap;
            word-break: normal;
            word-wrap: normal;
            tab-size: 4;
        }
        
        .line-number {
            display: inline-block;
            width: 3em;
            color: #6a737d;
            text-align: right;
            padding-right: 1em;
            user-select: none;
        }
        
        .doc-comment {
            background-color: #fffdf7;
            padding: 10px;
            border-left: 3px solid #f9c270;
            margin: 1em 0;
        }
        
        /* Navigation */
        .nav {
            background-color: #24292e;
            padding: 10px 20px;
            margin-bottom: 20px;
            border-radius: 5px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .nav-title {
            color: white;
            font-weight: bold;
            font-size: 1.2em;
        }
        
        .nav-links {
            display: flex;
            gap: 20px;
        }
        
        .nav-links a {
            color: #ffffff;
            text-decoration: none;
        }
        
        .nav-links a:hover {
            text-decoration: underline;
        }
        
        /* Section styles */
        .section {
            margin-bottom: 2em;
        }
        
        .summary {
            margin-bottom: 1em;
            color: #586069;
        }
        
        /* Card styles for types, funcs, etc. */
        .card {
            background-color: white;
            border: 1px solid #e1e4e8;
            border-radius: 3px;
            margin-bottom: 1em;
            overflow: hidden;
        }
        
        .card-header {
            background-color: #f6f8fa;
            padding: 10px 15px;
            border-bottom: 1px solid #e1e4e8;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .card-title {
            font-weight: 600;
            margin: 0;
        }
        
        .card-body {
            padding: 15px;
        }
        
        /* Type-specific styles */
        .type-struct {
            border-left: 4px solid #2ecc71;
        }
        
        .type-interface {
            border-left: 4px solid #3498db;
        }
        
        .type-alias {
            border-left: 4px solid #9b59b6;
        }
        
        .type-other {
            border-left: 4px solid #95a5a6;
        }
        
        /* Syntax highlighting */
        .keyword {
            color: #cf222e;
            font-weight: bold;
        }
        
        .string {
            color: #0a3069;
        }
        
        .comment {
            color: #6a737d;
            font-style: italic;
        }
        
        .function {
            color: #6f42c1;
        }
        
        .constant {
            color: #0550ae;
            font-weight: bold;
        }
        
        .variable {
            color: #24292f;
        }
        
        /* Custom styles */
        {{.CustomCSS}}
    </style>
    {{end}}
</head>
<body>
    <div class="container">
        <div class="nav">
            <div class="nav-title">{{.Package.Name}}</div>
            <div class="nav-links">
                <a href="#overview">Overview</a>
                <a href="#types">Types</a>
                <a href="#functions">Functions</a>
                <a href="#constants">Constants</a>
                <a href="#variables">Variables</a>
            </div>
        </div>
        
        <section id="overview" class="section">
            <h1>Package {{.Package.Name}}</h1>
            {{if .Package.PackageDoc}}
                {{formatDoc .Package.PackageDoc}}
            {{else}}
                <p class="summary">No package documentation available.</p>
            {{end}}
            
            <h2>Imports</h2>
            {{if .Package.Imports}}
                <div class="card">
                    <div class="card-body">
                        <ul>
                            {{range .Package.Imports}}
                                <li>
                                    {{if .Alias}}{{.Alias}} {{end}}
                                    "{{.Path}}"
                                    {{if .Comment}} // {{.Comment}}{{end}}
                                </li>
                            {{end}}
                        </ul>
                    </div>
                </div>
            {{else}}
                <p class="summary">No imports.</p>
            {{end}}
        </section>
        
        <section id="types" class="section">
            <h2>Types</h2>
            {{if .Package.Types}}
                {{range .Package.Types}}
                    <div class="card {{typeKindClass .Kind}}">
                        <div class="card-header">
                            <h3 class="card-title">type {{.Name}}</h3>
                            <span>{{.Kind}}</span>
                        </div>
                        <div class="card-body">
                            {{if .Doc}}
                                {{formatDoc .Doc}}
                            {{end}}
                            {{formatCode .Code}}
                            
                            {{if eq .Kind "struct"}}
                                {{if .Fields}}
                                    <h4>Fields</h4>
                                    <table>
                                        <thead>
                                            <tr>
                                                <th>Name</th>
                                                <th>Type</th>
                                                <th>Tag</th>
                                                <th>Comment</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {{range .Fields}}
                                                <tr>
                                                    <td>{{if .Name}}{{.Name}}{{else}}<em>embedded</em>{{end}}</td>
                                                    <td>{{.Type}}</td>
                                                    <td><code>{{.Tag}}</code></td>
                                                    <td>{{.Comment}}</td>
                                                </tr>
                                            {{end}}
                                        </tbody>
                                    </table>
                                {{end}}
                            {{end}}
                            
                            {{if eq .Kind "interface"}}
                                {{if .InterfaceMethods}}
                                    <h4>Methods</h4>
                                    <table>
                                        <thead>
                                            <tr>
                                                <th>Name</th>
                                                <th>Signature</th>
                                                <th>Comment</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {{range .InterfaceMethods}}
                                                <tr>
                                                    <td>{{.Name}}</td>
                                                    <td>{{if .Signature}}{{.Signature}}{{else}}<em>embedded interface</em>{{end}}</td>
                                                    <td>{{.Comment}}</td>
                                                </tr>
                                            {{end}}
                                        </tbody>
                                    </table>
                                {{end}}
                            {{end}}
                        </div>
                    </div>
                {{end}}
            {{else}}
                <p class="summary">No types defined in this package.</p>
            {{end}}
        </section>
        
        <section id="functions" class="section">
            <h2>Functions and Methods</h2>
            {{if .Package.Functions}}
                {{range .Package.Functions}}
                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">
                                {{if .Receiver}}
                                    func ({{if .Receiver.Name}}{{.Receiver.Name}} {{end}}{{.Receiver.Type}}) {{.Name}}
                                {{else}}
                                    func {{.Name}}
                                {{end}}
                            </h3>
                        </div>
                        <div class="card-body">
                            {{if .Doc}}
                                {{formatDoc .Doc}}
                            {{end}}
                            {{formatCode .Code}}
                        </div>
                    </div>
                {{end}}
            {{else}}
                <p class="summary">No functions defined in this package.</p>
            {{end}}
        </section>
        
        <section id="constants" class="section">
            <h2>Constants</h2>
            {{if .Package.Constants}}
                <div class="card">
                    <div class="card-body">
                        <table>
                            <thead>
                                <tr>
                                    <th>Name</th>
                                    <th>Type</th>
                                    <th>Value</th>
                                    <th>Comment</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range .Package.Constants}}
                                    <tr>
                                        <td>{{.Name}}</td>
                                        <td>{{.Type}}</td>
                                        <td><code>{{.Value}}</code></td>
                                        <td>{{.Comment}}</td>
                                    </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                </div>
            {{else}}
                <p class="summary">No constants defined in this package.</p>
            {{end}}
        </section>
        
        <section id="variables" class="section">
            <h2>Variables</h2>
            {{if .Package.Variables}}
                <div class="card">
                    <div class="card-body">
                        <table>
                            <thead>
                                <tr>
                                    <th>Name</th>
                                    <th>Type</th>
                                    <th>Value</th>
                                    <th>Comment</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range .Package.Variables}}
                                    <tr>
                                        <td>{{.Name}}</td>
                                        <td>{{.Type}}</td>
                                        <td><code>{{.Value}}</code></td>
                                        <td>{{.Comment}}</td>
                                    </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                </div>
            {{else}}
                <p class="summary">No variables defined in this package.</p>
            {{end}}
        </section>
        
        <footer>
            <p>Generated by Go-Tree HTML Generator</p>
        </footer>
    </div>
</body>
</html>`
