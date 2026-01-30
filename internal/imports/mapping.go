package imports

// moduleToPackage maps Python import names to their pip package names.
// This is needed when the import name differs from the package name.
var moduleToPackage = map[string]string{
	// Image Processing
	"PIL":       "Pillow",
	"cv2":       "opencv-python",
	"skimage":   "scikit-image",

	// Machine Learning / Data Science
	"sklearn":   "scikit-learn",
	"tensorflow": "tensorflow",
	"tf":        "tensorflow",
	"torch":     "torch",
	"keras":     "keras",
	"xgboost":   "xgboost",
	"lightgbm":  "lightgbm",
	"catboost":  "catboost",

	// Data Manipulation
	"numpy":     "numpy",
	"np":        "numpy",       // Common alias, though import np doesn't work
	"pandas":    "pandas",
	"pd":        "pandas",      // Common alias
	"scipy":     "scipy",
	"sympy":     "sympy",
	"statsmodels": "statsmodels",
	"pyarrow":   "pyarrow",
	"polars":    "polars",

	// Web Scraping / HTTP
	"bs4":       "beautifulsoup4",
	"requests":  "requests",
	"httpx":     "httpx",
	"aiohttp":   "aiohttp",
	"urllib3":   "urllib3",
	"selenium":  "selenium",
	"scrapy":    "scrapy",
	"lxml":      "lxml",

	// Configuration / Environment
	"yaml":      "PyYAML",
	"dotenv":    "python-dotenv",
	"toml":      "toml",
	"environ":   "environ-config",
	"decouple":  "python-decouple",

	// Database
	"psycopg2":  "psycopg2-binary",
	"pymysql":   "PyMySQL",
	"pymongo":   "pymongo",
	"redis":     "redis",
	"sqlalchemy": "SQLAlchemy",
	"peewee":    "peewee",
	"motor":     "motor",
	"asyncpg":   "asyncpg",

	// Web Frameworks
	"flask":     "Flask",
	"fastapi":   "fastapi",
	"django":    "Django",
	"starlette": "starlette",
	"sanic":     "sanic",
	"bottle":    "bottle",
	"tornado":   "tornado",
	"uvicorn":   "uvicorn",
	"gunicorn":  "gunicorn",
	"pydantic":  "pydantic",

	// Testing
	"pytest":    "pytest",
	"mock":      "mock",
	"faker":     "Faker",
	"hypothesis": "hypothesis",
	"responses": "responses",
	"httpretty": "httpretty",
	"vcrpy":     "vcrpy",

	// CLI / Terminal
	"click":     "click",
	"typer":     "typer",
	"rich":      "rich",
	"colorama":  "colorama",
	"tqdm":      "tqdm",
	"tabulate":  "tabulate",
	"fire":      "fire",

	// Async
	"trio":      "trio",
	"anyio":     "anyio",
	"gevent":    "gevent",
	"eventlet":  "eventlet",
	"celery":    "celery",

	// Serialization
	"msgpack":   "msgpack",
	"orjson":    "orjson",
	"ujson":     "ujson",
	"simplejson": "simplejson",
	"protobuf":  "protobuf",
	"avro":      "avro-python3",

	// Cryptography / Security
	"cryptography": "cryptography",
	"nacl":      "PyNaCl",
	"jwt":       "PyJWT",
	"passlib":   "passlib",
	"bcrypt":    "bcrypt",
	"paramiko":  "paramiko",

	// Cloud / AWS
	"boto3":     "boto3",
	"botocore":  "botocore",
	"google":    "google-cloud",
	"azure":     "azure",

	// Visualization
	"matplotlib": "matplotlib",
	"plt":       "matplotlib",  // Common alias
	"seaborn":   "seaborn",
	"sns":       "seaborn",     // Common alias
	"plotly":    "plotly",
	"bokeh":     "bokeh",
	"altair":    "altair",

	// NLP
	"nltk":      "nltk",
	"spacy":     "spacy",
	"transformers": "transformers",
	"gensim":    "gensim",
	"textblob":  "textblob",

	// Date/Time
	"dateutil":  "python-dateutil",
	"arrow":     "arrow",
	"pendulum":  "pendulum",
	"pytz":      "pytz",

	// Utilities
	"attr":      "attrs",
	"attrs":     "attrs",
	"more_itertools": "more-itertools",
	"toolz":     "toolz",
	"cytoolz":   "cytoolz",
	"boltons":   "boltons",
	"sh":        "sh",
	"plumbum":   "plumbum",
	"invoke":    "invoke",
	"fabric":    "fabric",

	// Logging / Monitoring
	"loguru":    "loguru",
	"structlog": "structlog",
	"sentry_sdk": "sentry-sdk",

	// Validation
	"marshmallow": "marshmallow",
	"cerberus":  "Cerberus",
	"voluptuous": "voluptuous",
	"jsonschema": "jsonschema",

	// API
	"graphene":  "graphene",
	"strawberry": "strawberry-graphql",
	"grpc":      "grpcio",

	// Jupyter / Notebooks
	"IPython":   "ipython",
	"ipywidgets": "ipywidgets",
	"nbformat":  "nbformat",

	// Misc
	"Pillow":    "Pillow",
	"networkx":  "networkx",
	"igraph":    "python-igraph",
	"shapely":   "shapely",
	"fiona":     "Fiona",
	"rasterio":  "rasterio",
	"geopandas": "geopandas",
	"folium":    "folium",
	"pygments":  "Pygments",
	"jinja2":    "Jinja2",
	"mako":      "Mako",
	"chardet":   "chardet",
	"ftfy":      "ftfy",
	"unidecode": "Unidecode",
	"emoji":     "emoji",
}

// GetPackageName returns the pip package name for a given Python module.
// If no mapping exists, the module name is returned as-is (works for most packages).
func GetPackageName(module string) string {
	if pkg, ok := moduleToPackage[module]; ok {
		return pkg
	}
	return module
}
