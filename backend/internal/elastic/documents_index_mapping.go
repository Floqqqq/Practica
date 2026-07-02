package elastic

const documentsIndexMapping = `{
  "settings": {
    "analysis": {
      "filter": {
        "russian_stop": {
          "type": "stop",
          "stopwords": "_russian_"
        },
        "russian_stemmer": {
          "type": "stemmer",
          "language": "russian"
        }
      },
      "analyzer": {
        "analysis_ru": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "russian_stop",
            "russian_stemmer"
          ]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "chunk_id": {
        "type": "keyword"
      },
      "document_id": {
        "type": "keyword"
      },
      "file_name": {
        "type": "keyword"
      },
      "page_number": {
        "type": "integer"
      },
      "chunk_index": {
        "type": "integer"
      },
      "text": {
        "type": "text",
        "analyzer": "analysis_ru",
        "search_analyzer": "analysis_ru"
      },
      "start_offset": {
        "type": "integer"
      },
      "end_offset": {
        "type": "integer"
      },
      "chars_count": {
        "type": "integer"
      },
      "indexed_at": {
        "type": "date"
      }
    }
  }
}`
