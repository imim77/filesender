package main

type Config struct {
	Host string
	Port string
}

type Server struct {
	cfg *Config
}
