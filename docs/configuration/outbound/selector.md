### Structure

```json
{
  "type": "selector",
  "tag": "select",
  
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
  "default": "proxy-c",
  "interrupt_exist_connections": false
}
```

!!! error ""

    The selector can only be controlled through the [Clash API](/configuration/experimental#clash-api-fields) currently.

!!! note ""

    You can ignore the JSON Array [] tag when the content is only one item

### Fields

#### outbounds

List of outbound tags to select.

#### providers

List of providers tags to select.

#### includes

List of regular expression used to match tag of outbounds contained by providers which can be appended.

#### excludes

Match tag of outbounds contained by providers which cannot be appended.

#### types

Match type of outbounds contained by providers which cannot be appended.

#### default

The default outbound tag. The first outbound will be used if empty.

#### interrupt_exist_connections

Interrupt existing connections when the selected outbound has changed.

Only inbound connections are affected by this setting, internal connections will always be interrupted.
