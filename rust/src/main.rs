use actix_web::{get, web, App, HttpRequest, HttpResponse, HttpServer, Responder};
use dotenv::dotenv;
use jsonwebtoken::{decode, Algorithm, DecodingKey, Validation};
use serde::{Deserialize, Serialize};
use sqlx::mysql::{MySqlPool, MySqlPoolOptions};

#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    iat: usize,
    email: String,
}

#[derive(Debug, Deserialize, Serialize, sqlx::FromRow)]
#[allow(non_snake_case)]
pub struct User {
    pub email: String,
    pub first: Option<String>,
    pub last: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub age: Option<i32>,
}

pub struct AppState {
    db: MySqlPool,
    jwt_secret: String,
}

#[get("/")]
async fn get_user(req: HttpRequest, data: web::Data<AppState>) -> impl Responder {
    let validation = Validation::new(Algorithm::HS256);
    let mut auth_hdr: &str = req
        .headers()
        .get(actix_web::http::header::AUTHORIZATION)
        .unwrap()
        .to_str()
        .unwrap();
    auth_hdr = &auth_hdr.strip_prefix("Bearer ").unwrap();
    let token = match decode::<Claims>(
        &auth_hdr,
        &DecodingKey::from_secret(data.jwt_secret.as_ref()),
        &validation,
    ) {
        Ok(c) => c,
        Err(e) => {
            eprintln!("Application error: {e}");
            return HttpResponse::InternalServerError().into();
        }
    };

    let email: String = token.claims.email;
    let query_result = sqlx::query_as!(User, r#"SELECT *  FROM users WHERE email = ?"#, email)
        .fetch_one(&data.db)
        .await;

    match query_result {
        Ok(user) => {
            let user_response = serde_json::json!(user);
            return HttpResponse::Ok().json(user_response);
        }
        Err(sqlx::Error::RowNotFound) => {
            return HttpResponse::NotFound().into();
        }
        Err(_e) => {
            return HttpResponse::InternalServerError().into();
        }
    };
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();
    env_logger::init();
    let database_url = std::env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    let pool = match MySqlPoolOptions::new()
        .max_connections(300)
        .connect(&database_url)
        .await
    {
        Ok(pool) => pool,
        Err(err) => {
            println!("🔥 Failed to connect to the database: {:?}", err);
            std::process::exit(1);
        }
    };

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(AppState {
                db: pool.clone(),
                jwt_secret: std::env::var("JWT_SECRET").unwrap(),
            }))
            .service(get_user)
    })
        .bind(("127.0.0.1", 3001))?
        .run()
        .await
}
