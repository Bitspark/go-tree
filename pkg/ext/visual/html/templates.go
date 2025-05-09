// Package html provides HTML visualization for Go modules.
package html

// BaseTemplate is the basic HTML template structure
const BaseTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        :root {
            --primary-color: #00ADD8;
            --secondary-color: #5DC9E2;
            --background-color: #FFFFFF;
            --text-color: #333333;
            --highlight-color: #FFF3BF;
            --border-color: #E1E4E8;
            --heading-color: #0A1922;
            --symbol-fn-color: #6A3D9A;
            --symbol-type-color: #1F78B4;
            --symbol-var-color: #33A02C;
            --symbol-const-color: #E31A1C;
            --symbol-field-color: #FF7F00;
            --symbol-pkg-color: #A6CEE3;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
            line-height: 1.5;
            color: var(--text-color);
            background-color: var(--background-color);
            margin: 0;
            padding: 20px;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
        }

        h1, h2, h3, h4 {
            color: var(--heading-color);
        }

        h1 {
            border-bottom: 2px solid var(--primary-color);
            padding-bottom: 0.5rem;
        }

        h2 {
            border-bottom: 1px solid var(--border-color);
            padding-bottom: 0.3rem;
        }

        .module-info {
            background-color: #F6F8FA;
            border: 1px solid var(--border-color);
            border-radius: 4px;
            padding: 15px;
            margin-bottom: 20px;
        }

        .package {
            margin-bottom: 30px;
            border: 1px solid var(--border-color);
            border-radius: 4px;
            padding: 15px;
        }

        .package-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .symbol {
            margin: 10px 0;
            padding: 10px;
            border-radius: 4px;
            border-left: 3px solid var(--secondary-color);
            background-color: #F6F8FA;
        }

        .symbol-name {
            font-weight: bold;
        }

        .symbol-type {
            color: #666;
            font-family: SFMono-Regular, Consolas, Liberation Mono, Menlo, monospace;
        }

        .symbol-fn {
            border-left-color: var(--symbol-fn-color);
        }

        .symbol-type {
            border-left-color: var(--symbol-type-color);
        }

        .symbol-var {
            border-left-color: var(--symbol-var-color);
        }

        .symbol-const {
            border-left-color: var(--symbol-const-color);
        }

        .symbol-field {
            border-left-color: var(--symbol-field-color);
        }

        .symbol-pkg {
            border-left-color: var(--symbol-pkg-color);
        }

        .highlight {
            background-color: var(--highlight-color);
        }

        .code {
            font-family: SFMono-Regular, Consolas, Liberation Mono, Menlo, monospace;
            padding: 2px 4px;
            background-color: #F0F0F0;
            border-radius: 3px;
        }

        .tag {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 12px;
            font-weight: bold;
            margin-right: 5px;
        }

        .tag-exported {
            background-color: #E3F2FD;
            color: #0D47A1;
        }

        .tag-private {
            background-color: #EEEEEE;
            color: #616161;
        }

        .tag-interface {
            background-color: #E8F5E9;
            color: #1B5E20;
        }

        .tag-struct {
            background-color: #FFF3E0;
            color: #E65100;
        }

        .type-info {
            margin-top: 4px;
            padding: 4px 8px;
            background-color: #F8F9FA;
            border-radius: 4px;
            font-family: SFMono-Regular, Consolas, Liberation Mono, Menlo, monospace;
            font-size: 0.9em;
        }

        .references {
            margin-top: 10px;
            font-size: 0.9em;
        }

        .references-title {
            font-weight: bold;
            margin-bottom: 5px;
        }

        .reference {
            padding: 2px 4px;
            margin: 2px 0;
            border-radius: 3px;
            background-color: #F0F0F0;
        }

        table {
            width: 100%;
            border-collapse: collapse;
        }

        table, th, td {
            border: 1px solid var(--border-color);
        }

        th, td {
            padding: 8px 12px;
            text-align: left;
        }

        th {
            background-color: #F6F8FA;
        }

        .relationship-graph {
            margin-top: 20px;
            border: 1px solid var(--border-color);
            border-radius: 4px;
            padding: 15px;
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --background-color: #0D1117;
                --text-color: #C9D1D9;
                --border-color: #30363D;
                --heading-color: #FFFFFF;
                --highlight-color: #2D333B;
            }

            .module-info, .symbol, .package {
                background-color: #161B22;
                border-color: var(--border-color);
            }

            .code, .type-info, .reference {
                background-color: #1A1A1A;
            }

            table, th, td {
                border-color: var(--border-color);
            }

            th {
                background-color: #161B22;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        <div class="module-info">
            <div><strong>Module Path:</strong> {{.ModulePath}}</div>
            <div><strong>Go Version:</strong> {{.GoVersion}}</div>
            <div><strong>Packages:</strong> {{.PackageCount}}</div>
        </div>

        {{.Content}}
    </div>
</body>
</html>
`
