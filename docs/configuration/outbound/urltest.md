### Structure

```json
{
  "type": "urltest",
  "tag": "auto",
  
  "outbounds": [
    "proxy-a",
    "proxy-b",
    "proxy-c"
  ],
  "providers": [
    "provider-a",
    "provider-b",
    "provider-c",
  ],
  "includes": [
    "^HK\\..+",
    "^TW\\..+",
    "^SG\\..+",
  ],
  "excludes": "^JP\\..+",
  "types": [
    "shadowsocks",
    "vmess",
    "vless",
  ],
  "url": "https://www.gstatic.com/generate_204",
  "interval": "1m",
  "tolerance": 50,
  "interrupt_exist_connections": false
}
```

!!! note ""

    You can ignore the JSON Array [] tag when the content is only one item

### Fields

#### outbounds

List of outbound tags to test.

#### providers

List of providers tags to select.

#### includes

List of regular expression used to match tag of outbounds contained by providers which can be appended.

#### excludes

Match tag of outbounds contained by providers which cannot be appended.

#### types

Match type of outbounds contained by providers which cannot be appended.

#### url

The URL to test. `https://www.gstatic.com/generate_204` will be used if empty.

#### interval

The test interval. `1m` will be used if empty.

#### tolerance

The test tolerance in milliseconds. `50` will be used if empty.

#### interrupt_exist_connections

Interrupt existing connections when the selected outbound has changed.

Only inbound connections are affected by this setting, internal connections will always be interrupted.
