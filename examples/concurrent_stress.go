/*
 * concurrent_stress.go
 *
 * Copyright © 2013, S.Çağlar Onur
 *
 * Authors:
 * S.Çağlar Onur <caglar@10ur.org>
 *
 * This library is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 2, as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
 */

package main

import (
	"flag"
	"github.com/caglar10ur/gologger"
	"github.com/caglar10ur/lxc"
	"runtime"
	"strconv"
	"sync"
)

var (
	iteration int
	count     int
	template  string
	debug     bool
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&template, "template", "busybox", "Template to use")
	flag.IntVar(&count, "count", 10, "Number of operations to run concurrently")
	flag.IntVar(&iteration, "iteration", 1, "Number times to run the test")
	flag.BoolVar(&debug, "debug", false, "Flag to control debug output")
	flag.Parse()
}

func main() {
	log := logger.New(nil)
	if debug {
		log.SetLogLevel(logger.Debug)
	}

	log.Debugf("Using %d GOMAXPROCS", runtime.NumCPU())

	var wg sync.WaitGroup

	for i := 0; i < iteration; i++ {
		log.Debugf("-- ITERATION %d --", i+1)
		for _, mode := range []string{"CREATE", "START", "STOP", "DESTROY"} {
			log.Debugf("\t-- %s --", mode)
			for j := 0; j < count; j++ {
				wg.Add(1)
				go func(i int, mode string) {
					c, err := lxc.NewContainer(strconv.Itoa(i))
					if err != nil {
						log.Fatalf("ERROR: %s\n", err.Error())
					}
					defer lxc.PutContainer(c)

					if mode == "CREATE" {
						log.Debugf("\t\tCreating the container (%d)...\n", i)
						if err := c.Create(template); err != nil {
							log.Errorf("\t\t\tERROR: %s\n", err.Error())
						}
					} else if mode == "START" {
						c.SetDaemonize()
						log.Debugf("\t\tStarting the container (%d)...\n", i)
						if err := c.Start(false); err != nil {
							log.Errorf("\t\t\tERROR: %s\n", err.Error())
						}
					} else if mode == "STOP" {
						log.Debugf("\t\tStoping the container (%d)...\n", i)
						if err := c.Stop(); err != nil {
							log.Errorf("\t\t\tERROR: %s\n", err.Error())
						}
					} else if mode == "DESTROY" {
						log.Debugf("\t\tDestroying the container (%d)...\n", i)
						if err := c.Destroy(); err != nil {
							log.Errorf("\t\t\tERROR: %s\n", err.Error())
						}
					}
					wg.Done()
				}(j, mode)
			}
			wg.Wait()
		}
	}
}
