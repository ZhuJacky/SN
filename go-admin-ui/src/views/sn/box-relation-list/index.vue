<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
          <el-form ref="printConf" :model="printConf" :rules="rules" :inline="true" label-width="120px">
            <el-form-item label="打印数量" prop="Number">
              <el-input-number :disabled="noEdit" readonly="noEdit" v-model="printConf.Number" controls-position="right" :min="0" />
            </el-form-item>
            <el-form-item label="编号" prop="sortCode">
            <el-input
              v-model="printConf.sortCode"
              placeholder="请输入编号"
              clearable
              size="small"
            />
          </el-form-item>
          </el-form>
        <el-form ref="queryForm" :model="queryParams" :inline="true" label-width="120px">
          <el-form-item label="箱号" prop="BoxId">
            <el-input
              v-model="queryParams.BoxId"
              placeholder=""
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
          </el-form-item>
        </el-form>

        <el-row :gutter="10" class="mb8">
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-microphone"
              size="mini"
              @click="warningAudio"
            ></el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-microphone"
              size="mini"
              @click="addBox"
            >测试扫码功能</el-button>
          </el-col>
        </el-row>

        <el-table v-loading="loading" :data="boxRelationList" border >
          <el-table-column label="序号" width="120" align="center" prop="BoxRelationId" />
          <el-table-column label="箱号" width="120" align="center" prop="BoxId" />
          <el-table-column label="SN号" align="center" prop="SNCode" />
          <el-table-column label="装箱数量" align="center" prop="BoxSum" />
          <el-table-column label="扫码枪IP" width="120" align="center" prop="ScanSource" />
          <el-table-column label="创建时间" align="center" prop="createdAt" width="155">
            <template slot-scope="scope">
              <span>{{ parseTime(scope.row.createdAt) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <pagination
          v-show="total>0"
          :total="total"
          :page.sync="queryParams.pageIndex"
          :limit.sync="queryParams.pageSize"
          @pagination="getList"
        />
      </el-card>


        <div ref="printDiv" style="display:none;" >
                <div style="font-size:10px;margin-top:7px;">
                    箱号：{{printItem.BoxId}}
                </div>
                <div style="font-size:10px;margin-top:7px;word-wrap:break-word">
                    SN列表：{{}}
                </div>
       </div>
    </template>
  </BasicLayout>
</template>

<script>
import { listPost,packBox } from '@/api/sn/box-relation-info'
import { formatJson } from '@/utils'
import { print } from '@/api/sn/sn-info'
import moment from 'moment'
import {clean,printparamsJsonArray,PTK_OpenUSBPort, PTK_CloseUSBPort, PTK_ClearBuffer,PTK_SetDarkness,PTK_SetPrintSpeed,PTK_SetDirection,PTK_SetLabelHeight,PTK_SetLabelWidth,PTK_DrawText_TrueType,PTK_PrintLabel,PTK_DrawBar2D_DATAMATRIX} from '@/utils/POSTEK'
import bwipjs from 'bwip-js'
import request from '@/utils/request'

export default {
  name: 'BoxRelationInfoManage',
  data() {
    return {
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      // 非多个禁用
      multiple: true,
      // 总条数
      total: 0,
      // 岗位表格数据
      boxRelationList: [],
      // 弹出层标题
      title: '',
      // 是否显示弹出层
      open: false,
      // 状态数据字典
      statusOptions: [],
      // 查询参数
      queryParams: {
        pageIndex: 1,
        pageSize: 100,
        postCode: undefined,
        postName: undefined,
        status: undefined
      },
      boxData: {
        snCode: undefined,
        status: 3,
        scanSource: '666'
      },
      printItem:{
        SNCodeList:[],
      },
      printConf:{Number:1},
              goodCodeList:[  
　　　　 　　{uid: 1, brandName: '苹果',goodName: 'iphone13 Pro', code: '11112222333',price: 900, unit:'个', skuName: '8+526G', num: 1},
            {uid: 2, brandName: '阿里',goodName: 'iphone13 Pro', code: '11112222333',price: 900, unit:'个', skuName: '8+526G', num: 1},
        ],
      // 表单参数
      form: {
      },
      // 表单校验
      rules: {
        sortCode: [
          { required: true, message: '编号不能为空', trigger: 'blur' }
        ],
      }
    }
  },
  created() {
    this.getList()

    window.document.onkeypress = (e) => {
      if (window.event) { // IE
        this.nextCode = e.keyCode
      } else if (e.which) { // Netscape/Firefox/Opera
        this.nextCode = e.which
      }
 
      if (e.which === 13) { // 键盘回车事件
        if (this.code.length < 3) return // 扫码枪的速度很快，手动输入的时间不会让code的长度大于2，所以这里不会对扫码枪有效
        console.log('扫码结束。')
        console.log('条形码：', this.code)
        this.parseQRCode(this.code) // 获取到扫码枪输入的内容，做别的操作
        this.lastCode = ''
        this.lastTime = ''
        return
      }
 
      this.nextTime = new Date().getTime()
      if (!this.lastTime && !this.lastCode) {
        this.code = '' // 清空上次的条形码
        this.code += e.key
        console.log('扫码开始---', this.code)
      }
      if (this.lastCode && this.lastTime && this.nextTime - this.lastTime > 500) { // 当扫码前有keypress事件时,防止首字缺失
        this.code = e.key
        console.log('防止首字缺失。。。', this.code)
      } else if (this.lastCode && this.lastTime) {
        this.code += e.key
        console.log('扫码中。。。', this.code)
      }
      this.lastCode = this.nextCode
      this.lastTime = this.nextTime
    }
  },
  methods: {

    /** 查询装箱列表 */
    getList() {
      this.loading = true
      listPost(this.queryParams).then(response => {
        this.boxRelationList = response.data.list
        this.total = response.data.count
        this.loading = false
                
      })
    },
    /** 搜索按钮操作 */
    handleQuery() {
      this.queryParams.pageIndex = 1
      this.getList()
    },
    
    /** 当SN码有异常时，触发告警声音 001 */
    warningAudio() {
      this.audio = new Audio()
      this.audio.src  = "http://159.75.213.231:8000/static/audios/do_wrong.mp3"
      this.audio.play()
    },
    parseQRCode(code) {
      if (code.length === 13) {
        // 处理自己的逻辑
        //alert('条码合法：' + code)
        this.$emit('barCode', code) //通知父组件
         this.addBox(code)
      } else if (code.length === 23) {
        console.log('B类条码:' + code)
      } else if (code.length === 0) {
        console.log('请输入条码！')
      } else {
        this.addBox(code)
        //alert('条码不合法：' + code)
      }
      this.codeValue = code
      // 发送网络请求
    },
    addBox(code) { //开始装箱
      
      //alert('开始装箱：' + code)

      //code = '8930500A000007'
      //alert(this.boxData.snCode)
  
      packBox(this.boxData, code).then(response => {
          if (response.code === 200) {              
              let Status = response.data.Status
              if(Status==1) {
                this.warningAudio()
                this.msgError(response.msg)
              } else if(Status==2) {
                this.warningAudio()
                this.msgError(response.msg)
              } else if(Status==3) {
                this.warningAudio()
                this.msgError(response.msg)
              } else if(Status==0) { //装箱成功
                this.queryParams.BoxId = response.data.BoxId;//填充查询条件

                this.getList()

                
                
                

              } else if(Status==4) { //装满一箱，调用打印机执行打印动作
                this.queryParams.BoxId = response.data.BoxId;//填充查询条件

                for (var i=0;i<response.data.BoxSNCodeList.length;i++)
                {
                  this.printItem.BoxId=response.data.BoxId
                  this.printItem.sortCode=this.printConf.sortCode
                  this.printItem.BatchCode=response.data.BoxSNCodeList[i].BatchCode
                  this.printItem.SNCodeList[i]=response.data.BoxSNCodeList[i].SNCode
                }
                this.doPrint()
                this.getList()
                
                
                //this.doPrint()                
              }

          } else {
              this.msgError(response.msg)
          }
      })
    },
    doPrint() {

      PTK_OpenUSBPort(255);//打开打印机USB端口
        this.PrintContent();         //打印内容
		    PTK_CloseUSBPort();  //关闭USB端口
		    this.printing();          //请求数据并打印
    },

    PrintContent(){
      var mm=12;
     var printNum=this.printConf.Number; //打印份数
    
     var width=100*mm;//单列标签的宽度 （每张小标签的宽度）
		 PTK_ClearBuffer();     //*清空缓存
		 PTK_SetDarkness(20);   //设置打印黑度 取值范围 0-20
		 PTK_SetPrintSpeed(8);  //设置打印速度
		 PTK_SetDirection('B'); //设置打印方向
		 PTK_SetLabelHeight(100*mm,24,24,true); //*设置标签高度、间隙及偏移
		 PTK_SetLabelWidth(100*mm);//*设置标签宽度(底纸的宽度)，一定要准确，否则会导致打印内容位置不准确

    for (var i = 1; i < printNum+1; i++) {
			var col_cr=1
      
		  //PTK_DrawBar2D_QREx((col_cr-1)*width+10,10,0,5,1,0,8,"ss","博思得科技发展有限公司");//打印一个QR码（二维码） 				可根据版本号来固定二维码大小（注意：数据量超过版本所能容纳的量则打印失败） 版本号为0则根据内容自动生成二维码（大小固定）
      PTK_DrawBar2D_DATAMATRIX((col_cr-1)*width+30,50,0,0,0,10,this.printItem.BoxId)
      PTK_DrawText_TrueType((col_cr-1)*width+30,160,3*mm,0,"宋体",1,1200,0,0,0,"箱号"); //打印snCode
      PTK_DrawText_TrueType((col_cr-1)*width+120,160,3*mm,0,"Arial",1,1200,0,0,0,this.printItem.BoxId); //打印snCode

      PTK_DrawText_TrueType((col_cr-1)*width+30,210,3*mm,0,"宋体",1,1200,0,0,0,"编号"); //打印
      PTK_DrawText_TrueType((col_cr-1)*width+120,210,3*mm,0,"Arial",1,1200,0,0,0,this.printConf.sortCode); //打印

      PTK_DrawText_TrueType((col_cr-1)*width+30,260,3*mm,0,"宋体",1,1200,0,0,0,"批号"); //打印
      PTK_DrawText_TrueType((col_cr-1)*width+120,260,3*mm,0,"Arial",1,1200,0,0,0,this.printItem.BatchCode); //打印
      PTK_DrawText_TrueType((col_cr-1)*width+30,310,3*mm,0,"宋体",1,1200,0,0,0,"SN号"); //打印
      for (var i=0;i<this.printItem.SNCodeList.length;i++)
      {

           PTK_DrawText_TrueType((col_cr-1)*width+120,310+i*50,3*mm,0,"Arial",1,1200,0,0,0,this.printItem.SNCodeList[i]); //打印
      }

			console.log("-------打印第"+i+"行----------");
			PTK_PrintLabel(1,1); //打印，必须要执行这一行，否则不会打印
		  }
	},
 printing(){
	    	var data = {};
	    	data.reqParam = "1";
	      	data.printparams = JSON.stringify( printparamsJsonArray);
	     	//jQuery.support.cors = true;  //适用于IE浏览器
			clean(); //此函数必须使用
	    	var url = "http://127.0.0.1:888/postek/print";
        //alert(url);
      //return request({    url: url,method: 'post',data: data,dataType:"json"});
      print(url,data).then(response => {
        
        if (response.retval === '0') {
	      alert("aasvv");
	        		} else {
	        			alert("存在错误，返回结果："+response.msg);
	        		}
      })
	  },

  }
}
</script>
<style lang='scss' scoped>
    .dayinID{width: 237px;height: 155px;border: 1px solid #000;margin-top: 10px;}
</style>
