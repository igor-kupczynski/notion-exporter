# notion-exporter

Notion exporter is a simple tool to recursively "sync" notion pages to output directory

## Usage

```sh
./notion-exporter  -pages <pages> -token <token> -output <output-dir> 
```

For example:
```sh
./notion-exporter -pages e65ccf14e13b48d1a6d1b14cd84c4bed -token redacted -output output
```

Arguments:
* `-pages`: comma-separated list of page to export the data from.
* `-token`: copy the token from  `token_v2` cookie. Only required if the pages you want to export are not public.
* `-output`: output directory. ⚠️ the existing content of the directory will be wiped out, use with care.


## Credits

It relies on the [notion export feature](https://www.notion.so/Export-a-page-as-Markdown-69b6031dd9454022abed8e23a86b0e1e), and the unofficial [notion API go client](https://www.notion.so/Export-a-page-as-Markdown-69b6031dd9454022abed8e23a86b0e1e).

I'd love to switch to the official API at some point, but currently the markdown export is not implemented there and not all block types in pages are supported.
