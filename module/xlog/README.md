to use, write in main:
 
	defer xlog.Flush()
	xlog.Init(tmConfig.GetString("xlog_dir"), tmConfig.GetInt("ann_logger_level"))
	xlog.Info("start...")

default logfiles are:

	xlog.Info -> annlog.log
	xlog.Dbg -> annlog.dbg
	xlog.Err -> annlog.err

	err dump dup to *.log

view running log:

	tail -f ./<xlog_dir>/annlog.log
