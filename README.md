# cronjob-manager

cronjob を一つのリソースで複数管理するための controller です

## How to Use

次のように CustomResource を作成します

```
apiVersion: cronjobmanager.varu3.me/v1beta1
kind: CronJobManager
metadata:
  name: cronjobmanager-sample
spec:
  image: ubuntu:latest
  cronjobSettings:
    - name: hoge-runner
      schedule: "0,5,10,15,20,25,30,35,40,45,50,55 * * * *"
      command: ["echo", "hoge"]
      type: "general"
    - name: huga-runner
      schedule: "0,5,10,15,20,25,30,35,40,45,50,55 * * * *"
      command: ["echo", "huga"]
      type: "general"
```

- `image`: 全体で使うコンテナイメージを指定します
- `cronJobSettings.name`: cronjob の name を指定します
- `cronJobSettings.schedule`: cron 式でスケジュールを指定します
- `cronjobSettings.command`: 実行するコマンドを指定します

このファイルをデプロイすると以下のように cronjob が作られます

```
NAME                        SCHEDULE                                                  SUSPEND   ACTIVE   LAST SCHEDULE   AGE
hoge-runner                 0,5,10,15,20,25,30,35,40,45,50,55 * * * *                 False     0        54m             4d
huga-runner                 0,5,10,15,20,25,30,35,40,45,50,55 * * * *                 False     0        3h54m           4d
```

以上
