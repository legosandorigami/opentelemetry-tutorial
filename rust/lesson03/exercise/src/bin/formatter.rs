use axum::{extract::Query, routing::get, Router};
use tokio::net::TcpListener;
use std::collections::HashMap;

async fn format_handler(Query(params): Query<HashMap<String, String>>) -> String {
    match params.get("hello_to"){
        Some(hello_to) =>{
            format!("Hello, {}!", hello_to)
        },
        None => "Hello, !".to_string()
    }
}



#[tokio::main]
async fn main() {
    let app = Router::new().route("/format", get(format_handler));
    
    let listener = TcpListener::bind("0.0.0.0:8081").await.unwrap();
    
    axum::serve(listener, app).await.unwrap();
}