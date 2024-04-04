## ELK

```{image} ../img/elk.svg
:width: 200px
```

### Components

```{revealjs-fragments}
* Elasticsearch
* Logstash
* Kibana
* *beats
```

### Architecture

```{image} ../img/elk-slides1.png
:width: 700px
```

### Architecture
```{image} ../img/elk-slides2.png
:width: 700px
```

### Elasticsearch
```{revealjs-fragments}
* Index
* Mapping
* Document
* Field
```

### Search
```json
GET /_search
{
  "query": {
    "query_string": {
      "query": "(new york city) OR (big apple)",
      "default_field": "content"
    }
  }
}
```

### Elastic Index Shards
```{image} ../img/elk-slides3.png
:width: 700px
```

### Elastic Index Status
* Green
* Yellow
* Red

### Elastic ILM
```{image} ../img/elk-slides4.png
:width: 700px
```

### Elastic Nodes
```{revealjs-fragments}
* Master
* Data
* Ingest
```

### Logstash
```nginx
input {
  ...
}

filter {
  ...
}

output {
  ...
}
```

### Kibana
```{revealjs-fragments}
* Visualization
* Dashboards
* Search and Analytics
* Monitoring and Alerting
```
