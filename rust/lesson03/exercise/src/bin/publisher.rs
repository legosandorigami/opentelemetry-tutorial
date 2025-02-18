use axum::{extract::Query, routing::get, Router};
use tokio::net::TcpListener;
use std::collections::HashMap;

async fn publish_handler(Query(params): Query<HashMap<String, String>>) {
    if let Some(hello_str)  = params.get("hello_str"){
        println!("{}", hello_str);
    }
}



#[tokio::main]
async fn main() {
    let app = Router::new().route("/publish", get(publish_handler));
    
    let listener = TcpListener::bind("0.0.0.0:8082").await.unwrap();
    
    axum::serve(listener, app).await.unwrap();
}