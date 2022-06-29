# MySSLEE_QCloud

| backend（2台）      | checker（2台）                                           | check服务器 |
| ------------------- | -------------------------------------------------------- | ----------- |
| limit: 4, second: 5 | fast: 460, chan: 5, checked: 37k \| full: 40, chan: 103  | 60%/60%     |
| limit:5, second: 5  | fast: 500, chan: 6, checked: 40k \| full: 45, chan: 103  | 70%/70%     |
| limit:5, second: 5  | fast: 600, chan: 10, checked: 50k \| full: 50, chan: 103 | 80%/80%     |


