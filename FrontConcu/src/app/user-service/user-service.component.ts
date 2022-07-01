import { Injectable } from '@angular/core';
import { HttpClient} from '@angular/common/http';
import { Observable} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class UserService {


  private baseURL='http://localhost:3000';

  constructor(private http: HttpClient) { }

  createUser(tweet: Object): Observable<Object>{
    return this.http.post(`${this.baseURL}/new`, tweet);
  }

  getUserList(): Observable<any>{
    return this.http.get(`${this.baseURL}`);
  }
  
}