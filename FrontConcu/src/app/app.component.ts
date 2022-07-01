import { Component } from '@angular/core';
import { UserService } from './user-service/user-service.component';
import { Tweet } from './models/user';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {

  tweet: Tweet = new Tweet();
  constructor(private userService: UserService) { }
  userid: string = '';
  message: string = '';
  date: string = '';
  list1 = [{
    user_id: "ultima prueba 2",
    message: "Abimael",
    publish_date: "02-07-2022",
  }]

  ngOnInit(): void {
    this.userService.getUserList().
      subscribe(datos => {
        console.log(datos)
        this.list1.pop()
        if (datos != null) {
          for (var i = 0; i < datos.length; i++) {
            const a = String(datos[i].data.user_id)
            const b = String(datos[i].data.message)
            const c = String(datos[i].data.publish_date)
            const array = [a, b, c]
            this.list1.push({ user_id: a, message: b, publish_date: c })
          }
          console.log(this.list1)
        } else {
          null;
        }
      })

  }
  title = 'FrontConcu';


  CreateUser() {
    this.tweet.user_id=this.userid
    this.tweet.message=this.message
    this.tweet.publish_date=this.date
    console.log(this.tweet)
    this.userService.createUser(this.tweet)
      .subscribe(datos => console.log(datos), error => console.log(error));
    this.tweet = new Tweet()
  }
  GetUser() {
    this.userService.getUserList().
      subscribe(datos => {
        if (datos != null) {
          this.list1=[]
        for (var i = 0; i < datos.length; i++) {
          const a = String(datos[i].data.user_id)
          const b = String(datos[i].data.message)
          const c = String(datos[i].data.publish_date)
          const array = [a, b, c]
          this.list1.push({ user_id: a, message: b, publish_date: c })
        }
        console.log(this.list1)
      } else {
        null;
      }
      })
  }
}
