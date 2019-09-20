// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package notify

const reviewMailTemplate = `<h1>Hello {{ .Owner -}},</h1>

<p>
Cloudsweeper suspects that some of these resources could be unused. 
Please review, and either tag or delete them.
</p>

<p>If you want to delete a resource, either remove its
<b>cloudsweeper-whitelisted</b> tag or delete it manually.</p>

<p>
Conversely, if you want to keep a resource around for a longer time, then
add any of the following tags:
</p>

<ol>
	<li><b>cloudsweeper-whitelisted: (Resource stays around indefinitely.)</li>
	<li><b>cloudsweeper-expiry</b>: <i>YYYY-MM-DD</i> (Deletion occurs after the specified date.)</li>
	<li><b>cloudsweeper-lifetime</b>: days-<i>N</i> (Deletion occurs <i>N</i> days after resource was created.)</li>
</ol>

<p>
Read more about how Cloudsweeper works and how to better tag your resources 
<a href="https://agaridata.atlassian.net/wiki/spaces/EN/pages/808189987/Cloudsweeper">here</a>.
</p>

<h2>Old resources:</h2>
<p>
Resources marked <span style="background-color: #c9fc99;">in green</span> are whitelisted.
</p>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Instance type</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if and (even $i) (not (whitelisted $instance)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $instance }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $instance.Owner }}</td>
			<td>{{ productname $instance }}</td>
			<td>{{ rolename $instance }}</td>
			<td>{{ $instance.ID }}</td>
			<td>{{ instname $instance }}</td>
			<td>{{ $instance.InstanceType }}</td>
			<td>{{ $instance.Location }}</td>
			<td>{{ fdate $instance.CreationTime "2006-01-02" }} ({{ daysrunning $instance.CreationTime }})</td>
			<td>{{ accucost $instance }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Images) 0 }}
	<h3>Images</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
	<tr {{ if and (even $i) (not (whitelisted $image)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $image }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $image.Owner }}</td>
			<td>{{ productname $image }}</td>
			<td>{{ rolename $image }}</td>
			<td>{{ $image.ID }}</td>
			<td>{{ $image.Name }}</td>
			<td>{{ $image.Location }}</td>
			<td>{{ fdate $image.CreationTime "2006-01-02" }} ({{ daysrunning $image.CreationTime }})</td>
			<td>{{ accucost $image }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Volumes) 0 }}
	<h3>Volumes</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Attached to instance</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Volume type</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
	<tr {{ if and (even $i) (not (whitelisted $volume)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $volume }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $volume.Owner }}</td>
			<td>{{ productname $volume }}</td>
			<td>{{ rolename $volume }}</td>
			<td>{{ $volume.ID }}</td>
			<td>{{ $volume.SizeGB }} GB</td>
			<td>{{ $volume.Location }}</td>
			<td>{{ yesno $volume.Attached }}</td>
			<td>{{ fdate $volume.CreationTime "2006-01-02" }} ({{ daysrunning $volume.CreationTime }})</td>
			<td>{{ $volume.VolumeType }}</td>
			<td>{{ accucost $volume }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Snapshots) 0 }}
	<h3>Snapshots</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
	<tr {{ if and (even $i) (not (whitelisted $snapshot)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $snapshot }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $snapshot.Owner }}</td>
			<td>{{ productname $snapshot }}</td>
			<td>{{ rolename $snapshot }}</td>
			<td>{{ $snapshot.ID }}</td>
			<td>{{ $snapshot.SizeGB }} GB</td>
			<td>{{ $snapshot.Location }}</td>
			<td>{{ fdate $snapshot.CreationTime "2006-01-02" }} ({{ daysrunning $snapshot.CreationTime }})</td>
			<td>{{ accucost $snapshot }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Buckets) 0 }}
	<h3>Buckets</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Files</strong></th>
			<th><strong>Modified in < 6 months</strong></th>
			<th><strong>Monthly cost</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
	<tr {{ if and (even $i) (not (whitelisted $bucket)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $bucket }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $bucket.Owner }}</td>
			<td>{{ productname $bucket }}</td>
			<td>{{ rolename $bucket }}</td>
			<td>{{ $bucket.ID }}</td>
			<td>{{ printf "%.3f GB" $bucket.TotalSizeGB }}</td>
			<td>{{ $bucket.ObjectCount }}</td>
			<td>{{ modifiedInTheLast6Months $bucket.LastModified }}</td>
			<td>{{ printf "$%.3f" (bucketcost $bucket) }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal Cloudsweeper
</p>
`

const managerReviewMailTemplate = `<h1>Hello {{ .Owner -}},</h1>

<p>
This is a summary of all old/unused resources for your team.
</p>

<h2>Old resources:</h2>
<p>
Resources marked <span style="background-color: #c9fc99;">in green</span> are whitelisted.
</p>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Instance type</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if and (even $i) (not (whitelisted $instance)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $instance }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $instance.Owner }}</td>
			<td>{{ productname $instance }}</td>
			<td>{{ rolename $instance }}</td>
			<td>{{ $instance.ID }}</td>
			<td>{{ instname $instance }}</td>
			<td>{{ $instance.InstanceType }}</td>
			<td>{{ $instance.Location }}</td>
			<td>{{ fdate $instance.CreationTime "2006-01-02" }} ({{ daysrunning $instance.CreationTime }})</td>
			<td>{{ accucost $instance }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Images) 0 }}
	<h3>Images</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
	<tr {{ if and (even $i) (not (whitelisted $image)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $image }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $image.Owner }}</td>
			<td>{{ productname $image }}</td>
			<td>{{ rolename $image }}</td>
			<td>{{ $image.ID }}</td>
			<td>{{ $image.Name }}</td>
			<td>{{ $image.Location }}</td>
			<td>{{ fdate $image.CreationTime "2006-01-02" }} ({{ daysrunning $image.CreationTime }})</td>
			<td>{{ accucost $image }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Volumes) 0 }}
	<h3>Volumes</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Attached to instance</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Volume type</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
	<tr {{ if and (even $i) (not (whitelisted $volume)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $volume }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $volume.Owner }}</td>
			<td>{{ productname $volume }}</td>
			<td>{{ rolename $volume }}</td>
			<td>{{ $volume.ID }}</td>
			<td>{{ $volume.SizeGB }} GB</td>
			<td>{{ $volume.Location }}</td>
			<td>{{ yesno $volume.Attached }}</td>
			<td>{{ fdate $volume.CreationTime "2006-01-02" }} ({{ daysrunning $volume.CreationTime }})</td>
			<td>{{ $volume.VolumeType }}</td>
			<td>{{ accucost $volume }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Snapshots) 0 }}
	<h3>Snapshots</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
	<tr {{ if and (even $i) (not (whitelisted $snapshot)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $snapshot }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $snapshot.Owner }}</td>
			<td>{{ productname $snapshot }}</td>
			<td>{{ rolename $snapshot }}</td>
			<td>{{ $snapshot.ID }}</td>
			<td>{{ $snapshot.SizeGB }} GB</td>
			<td>{{ $snapshot.Location }}</td>
			<td>{{ fdate $snapshot.CreationTime "2006-01-02" }} ({{ daysrunning $snapshot.CreationTime }})</td>
			<td>{{ accucost $snapshot }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Buckets) 0 }}
	<h3>Buckets</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Files</strong></th>
			<th><strong>Modified in < 6 months</strong></th>
			<th><strong>Monthly cost</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
	<tr {{ if and (even $i) (not (whitelisted $bucket)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $bucket }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $bucket.Owner }}</td>
			<td>{{ productname $bucket }}</td>
			<td>{{ rolename $bucket }}</td>
			<td>{{ $bucket.ID }}</td>
			<td>{{ printf "%.3f GB" $bucket.TotalSizeGB }}</td>
			<td>{{ $bucket.ObjectCount }}</td>
			<td>{{ modifiedInTheLast6Months $bucket.LastModified }}</td>
			<td>{{ printf "$%.3f" (bucketcost $bucket) }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal Cloudsweeper
</p>
`

const totalReviewMailTemplate = `<h1>Hello {{ .Owner -}},</h1>

<p>
This is a summary of all old/unused resources for your org.
</p>

<h2>Old resources:</h2>
<p>
Resources marked <span style="background-color: #c9fc99;">in green</span> are whitelisted.
</p>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Instance type</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if and (even $i) (not (whitelisted $instance)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $instance }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $instance.Owner }}</td>
			<td>{{ productname $instance }}</td>
			<td>{{ rolename $instance }}</td>
			<td>{{ $instance.ID }}</td>
			<td>{{ instname $instance }}</td>
			<td>{{ $instance.InstanceType }}</td>
			<td>{{ $instance.Location }}</td>
			<td>{{ fdate $instance.CreationTime "2006-01-02" }} ({{ daysrunning $instance.CreationTime }})</td>
			<td>{{ accucost $instance }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Images) 0 }}
	<h3>Images</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
	<tr {{ if and (even $i) (not (whitelisted $image)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $image }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $image.Owner }}</td>
			<td>{{ productname $image }}</td>
			<td>{{ rolename $image }}</td>
			<td>{{ $image.ID }}</td>
			<td>{{ $image.Name }}</td>
			<td>{{ $image.Location }}</td>
			<td>{{ fdate $image.CreationTime "2006-01-02" }} ({{ daysrunning $image.CreationTime }})</td>
			<td>{{ accucost $image }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Volumes) 0 }}
	<h3>Volumes</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Attached to instance</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Volume type</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
	<tr {{ if and (even $i) (not (whitelisted $volume)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $volume }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $volume.Owner }}</td>
			<td>{{ productname $volume }}</td>
			<td>{{ rolename $volume }}</td>
			<td>{{ $volume.ID }}</td>
			<td>{{ $volume.SizeGB }} GB</td>
			<td>{{ $volume.Location }}</td>
			<td>{{ yesno $volume.Attached }}</td>
			<td>{{ fdate $volume.CreationTime "2006-01-02" }} ({{ daysrunning $volume.CreationTime }})</td>
			<td>{{ $volume.VolumeType }}</td>
			<td>{{ accucost $volume }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Snapshots) 0 }}
	<h3>Snapshots</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
	<tr {{ if and (even $i) (not (whitelisted $snapshot)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $snapshot }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $snapshot.Owner }}</td>
			<td>{{ productname $snapshot }}</td>
			<td>{{ rolename $snapshot }}</td>
			<td>{{ $snapshot.ID }}</td>
			<td>{{ $snapshot.SizeGB }} GB</td>
			<td>{{ $snapshot.Location }}</td>
			<td>{{ fdate $snapshot.CreationTime "2006-01-02" }} ({{ daysrunning $snapshot.CreationTime }})</td>
			<td>{{ accucost $snapshot }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Buckets) 0 }}
	<h3>Buckets</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Files</strong></th>
			<th><strong>Modified in < 6 months</strong></th>
			<th><strong>Monthly cost</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
	<tr {{ if and (even $i) (not (whitelisted $bucket)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $bucket }}style="background-color: #c9fc99;"{{ end }}>
			<td>{{ $bucket.Owner }}</td>
			<td>{{ productname $bucket }}</td>
			<td>{{ rolename $bucket }}</td>
			<td>{{ $bucket.ID }}</td>
			<td>{{ printf "%.3f GB" $bucket.TotalSizeGB }}</td>
			<td>{{ $bucket.ObjectCount }}</td>
			<td>{{ modifiedInTheLast6Months $bucket.LastModified }}</td>
			<td>{{ printf "$%.3f" (bucketcost $bucket) }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal Cloudsweeper
</p>
`

const deletionWarningTemplate = `<h1>Hello {{ .Owner -}},</h1>

<h2>Some of these resources will start being cleaned up 
in {{ timeUntilDelete .Instances .Images .Snapshots .Volumes .Buckets }} 
hours. To see the specific time(s), observe the deletion date column.</h2>

<p>
Unless you take action, the resources listed below will be cleaned up
from your account shortly.
</p>

<p>
If you want to save any of these resources from deletion,
add a tag with the key <b>cloudsweeper-whitelisted</b> to it.
</p>

<p>If you only want to keep the resource around for a while, first delete
its <b>cloudsweeper-delete-at</b> tag. Then add one of the following tags to it:</p>

<ol>
	<li><b>cloudsweeper-expiry</b>: <i>YYYY-MM-DD</i> (Deletion occurs after the specified date.)</li>
	<li><b>cloudsweeper-lifetime</b>: days-<i>N</i> (Deletion occurs <i>N</i> days after resource was created.)</li>
</ol>

<p>
Read more about how Cloudsweeper works and how to better tag your resources 
<a href="https://agaridata.atlassian.net/wiki/spaces/EN/pages/808189987/Cloudsweeper">here</a>.
</p>

<h2>Marked resources:</h2>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Instance type</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
			<th><strong>Deletion date</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $instance.Owner }}</td>
			<td>{{ productname $instance }}</td>
			<td>{{ rolename $instance }}</td>
			<td>{{ $instance.ID }}</td>
			<td>{{ instname $instance }}</td>
			<td>{{ $instance.InstanceType }}</td>
			<td>{{ $instance.Location }}</td>
			<td>{{ fdate $instance.CreationTime "2006-01-02" }} ({{ daysrunning $instance.CreationTime }})</td>
			<td>{{ accucost $instance }}</td>
			<td>{{ deletedate $instance "2006-01-02 (03:04 PM ET)" }}</td>	
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Images) 0 }}
	<h3>Images</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
			<th><strong>Deletion date</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $image.Owner }}</td>
			<td>{{ productname $image }}</td>
			<td>{{ rolename $image }}</td>
			<td>{{ $image.ID }}</td>
			<td>{{ $image.Name }}</td>
			<td>{{ $image.Location }}</td>
			<td>{{ fdate $image.CreationTime "2006-01-02" }} ({{ daysrunning $image.CreationTime }})</td>
			<td>{{ accucost $image }}</td>
			<td>{{ deletedate $image "2006-01-02 (03:04 PM ET)" }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Volumes) 0 }}
	<h3>Volumes</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Attached to instance</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Volume type</strong></th>
			<th><strong>Total cost</strong></th>
			<th><strong>Deletion date</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $volume.Owner }}</td>
			<td>{{ productname $volume }}</td>
			<td>{{ rolename $volume }}</td>
			<td>{{ $volume.ID }}</td>
			<td>{{ $volume.SizeGB }} GB</td>
			<td>{{ $volume.Location }}</td>
			<td>{{ yesno $volume.Attached }}</td>
			<td>{{ fdate $volume.CreationTime "2006-01-02" }} ({{ daysrunning $volume.CreationTime }})</td>
			<td>{{ $volume.VolumeType }}</td>
			<td>{{ accucost $volume }}</td>
			<td>{{ deletedate $volume "2006-01-02 (03:04 PM ET)" }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Snapshots) 0 }}
	<h3>Snapshots</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
			<th><strong>Deletion date</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $snapshot.Owner }}</td>
			<td>{{ productname $snapshot }}</td>
			<td>{{ rolename $snapshot }}</td>
			<td>{{ $snapshot.ID }}</td>
			<td>{{ $snapshot.SizeGB }} GB</td>
			<td>{{ $snapshot.Location }}</td>
			<td>{{ fdate $snapshot.CreationTime "2006-01-02" }} ({{ daysrunning $snapshot.CreationTime }})</td>
			<td>{{ accucost $snapshot }}</td>
			<td>{{ deletedate $snapshot "2006-01-02 (03:04 PM ET)" }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Buckets) 0 }}
	<h3>Buckets</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Files</strong></th>
			<th><strong>Modified in < 6 months</strong></th>
			<th><strong>Monthly cost</strong></th>
			<th><strong>Deletion date</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $bucket.Owner }}</td>
			<td>{{ productname $bucket }}</td>
			<td>{{ rolename $bucket }}</td>
			<td>{{ $bucket.ID }}</td>
			<td>{{ printf "%.3f GB" $bucket.TotalSizeGB }}</td>
			<td>{{ $bucket.ObjectCount }}</td>
			<td>{{ modifiedInTheLast6Months $bucket.LastModified }}</td>
			<td>{{ printf "$%.3f" (bucketcost $bucket) }}</td>
			<td>{{ deletedate $bucket "2006-01-02 (03:04 PM ET)" }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal Cloudsweeper
</p>
`

const markingDryRunTemplate = `<h1>Hello {{ .Owner -}},</h1>

<h2>These resources would have been marked for deletion if this was not a dry run.</h2>

<p>This was most likely done as a test. If you want to avoid seeing one of these
resources in a future email, add a Cloudsweeper tag to your resource.</p>

<p>
Read more about how Cloudsweeper works and how to better tag your resources 
<a href="https://agaridata.atlassian.net/wiki/spaces/EN/pages/808189987/Cloudsweeper">here</a>.
</p>

<h2>Marked resources:</h2>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Instance type</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $instance.Owner }}</td>
			<td>{{ productname $instance }}</td>
			<td>{{ rolename $instance }}</td>
			<td>{{ $instance.ID }}</td>
			<td>{{ instname $instance }}</td>
			<td>{{ $instance.InstanceType }}</td>
			<td>{{ $instance.Location }}</td>
			<td>{{ fdate $instance.CreationTime "2006-01-02" }} ({{ daysrunning $instance.CreationTime }})</td>
			<td>{{ accucost $instance }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Images) 0 }}
	<h3>Images</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $image.Owner }}</td>
			<td>{{ productname $image }}</td>
			<td>{{ rolename $image }}</td>
			<td>{{ $image.ID }}</td>
			<td>{{ $image.Name }}</td>
			<td>{{ $image.Location }}</td>
			<td>{{ fdate $image.CreationTime "2006-01-02" }} ({{ daysrunning $image.CreationTime }})</td>
			<td>{{ accucost $image }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Volumes) 0 }}
	<h3>Volumes</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Attached to instance</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Volume type</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $volume.Owner }}</td>
			<td>{{ productname $volume }}</td>
			<td>{{ rolename $volume }}</td>
			<td>{{ $volume.ID }}</td>
			<td>{{ $volume.SizeGB }} GB</td>
			<td>{{ $volume.Location }}</td>
			<td>{{ yesno $volume.Attached }}</td>
			<td>{{ fdate $volume.CreationTime "2006-01-02" }} ({{ daysrunning $volume.CreationTime }})</td>
			<td>{{ $volume.VolumeType }}</td>
			<td>{{ accucost $volume }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Snapshots) 0 }}
	<h3>Snapshots</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $snapshot.Owner }}</td>
			<td>{{ productname $snapshot }}</td>
			<td>{{ rolename $snapshot }}</td>
			<td>{{ $snapshot.ID }}</td>
			<td>{{ $snapshot.SizeGB }} GB</td>
			<td>{{ $snapshot.Location }}</td>
			<td>{{ fdate $snapshot.CreationTime "2006-01-02" }} ({{ daysrunning $snapshot.CreationTime }})</td>
			<td>{{ accucost $snapshot }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Buckets) 0 }}
	<h3>Buckets</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Product</strong></th>
			<th><strong>Role</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Files</strong></th>
			<th><strong>Modified in < 6 months</strong></th>
			<th><strong>Monthly cost</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $bucket.Owner }}</td>
			<td>{{ productname $bucket }}</td>
			<td>{{ rolename $bucket }}</td>
			<td>{{ $bucket.ID }}</td>
			<td>{{ printf "%.3f GB" $bucket.TotalSizeGB }}</td>
			<td>{{ $bucket.ObjectCount }}</td>
			<td>{{ modifiedInTheLast6Months $bucket.LastModified }}</td>
			<td>{{ printf "$%.3f" (bucketcost $bucket) }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal Cloudsweeper
</p>
`

const untaggedMailTemplate = `<h1>Hello {{ .Owner -}},</h1>

<p>
The following resources are not properly tagged. Tags are an important way of tracking the purposes and costs of resources. 
You are free to add other tags, but you should always include env, product, and role tags. 
You can read more about these tags <a href="https://agaridata.atlassian.net/wiki/spaces/EN/pages/4325488/Resource+Naming+and+Tagging+Requirements">here</a>.
</p>

<p>
Read more about how Cloudsweeper works and how to better tag your resources 
<a href="https://agaridata.atlassian.net/wiki/spaces/EN/pages/808189987/Cloudsweeper">here</a>.
</p>

<h2>Untagged resources:</h2>
<p><strong>Account ID:</strong> {{ .OwnerID }}</p>
<p>
Resources marked <span style="background-color: #c9fc99;">in green</span> are whitelisted.
</p>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Tags</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if and (even $i) (not (whitelisted $instance)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $instance }}style="background-color: #c9fc99;"{{ end }}>
			<td style="white-space: nowrap;">{{ $instance.Location }}</td>
			<td style="white-space: nowrap;">{{ $instance.ID }}</td>
			<td style="white-space: nowrap;">{{ daysrunning $instance.CreationTime }}</td>
			<td>
			{{ range $key, $val := $instance.Tags }}
			<span style="background-color: #d6d6d6; padding-top: 0.2em; padding-bottom: 0.2em; padding-left: 0.5em; padding-right: 0.5em; border-radius: 2em; margin-left: 0.1em; margin-right: 0.1em; margin-top:0.01em; margin-bottom: 0.01em; color: #000; display: inline-block;">{{ prettyTag $key $val }}</span>
			{{ end }}
			</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Images) 0 }}
	<h3>Images</h3>
	<p>Note that an image name is not the same as a Name tag</p>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Tags</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
	<tr {{ if and (even $i) (not (whitelisted $image)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $image }}style="background-color: #c9fc99;"{{ end }}>
			<td style="white-space: nowrap;">{{ $image.Location }}</td>
			<td style="white-space: nowrap;">{{ $image.ID }}</td>
			<td style="white-space: nowrap;">{{ daysrunning $image.CreationTime }}</td>
			<td>
			{{ range $key, $val := $image.Tags }}
			<span style="background-color: #d6d6d6; padding-top: 0.2em; padding-bottom: 0.2em; padding-left: 0.5em; padding-right: 0.5em; border-radius: 2em; margin-left: 0.1em; margin-right: 0.1em; margin-top:0.01em; margin-bottom: 0.01em; color: #000; display: inline-block;">{{ prettyTag $key $val }}</span>
			{{ end }}
			</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Volumes) 0 }}
	<h3>Volumes</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Tags</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
	<tr {{ if and (even $i) (not (whitelisted $volume)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $volume }}style="background-color: #c9fc99;"{{ end }}>
			<td style="white-space: nowrap;">{{ $volume.Location }}</td>
			<td style="white-space: nowrap;">{{ $volume.ID }}</td>
			<td style="white-space: nowrap;">{{ daysrunning $volume.CreationTime }}</td>
			<td>
			{{ range $key, $val := $volume.Tags }}
			<span style="background-color: #d6d6d6; padding-top: 0.2em; padding-bottom: 0.2em; padding-left: 0.5em; padding-right: 0.5em; border-radius: 2em; margin-left: 0.1em; margin-right: 0.1em; margin-top:0.01em; margin-bottom: 0.01em; color: #000; display: inline-block;">{{ prettyTag $key $val }}</span>
			{{ end }}
			</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Snapshots) 0 }}
	<h3>Snapshots</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Tags</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
	<tr {{ if and (even $i) (not (whitelisted $snapshot)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $snapshot }}style="background-color: #c9fc99;"{{ end }}>
			<td style="white-space: nowrap;">{{ $snapshot.Location }}</td>
			<td style="white-space: nowrap;">{{ $snapshot.ID }}</td>
			<td style="white-space: nowrap;">{{ daysrunning $snapshot.CreationTime }}</td>
			<td>
			{{ range $key, $val := $snapshot.Tags }}
			<span style="background-color: #d6d6d6; padding-top: 0.2em; padding-bottom: 0.2em; padding-left: 0.5em; padding-right: 0.5em; border-radius: 2em; margin-left: 0.1em; margin-right: 0.1em; margin-top:0.01em; margin-bottom: 0.01em; color: #000; display: inline-block;">{{ prettyTag $key $val }}</span>
			{{ end }}
			</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

{{ if gt (len .Buckets) 0 }}
	<h3>Buckets</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>ID</strong></th>
			<th><strong>Tags</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
	<tr {{ if and (even $i) (not (whitelisted $bucket)) }}style="background-color: #f2f2f2;"{{ else if whitelisted $bucket }}style="background-color: #c9fc99;"{{ end }}>
			<td style="white-space: nowrap;">{{ $bucket.ID }}</td>
			<td>
			{{ range $key, $val := $bucket.Tags }}
			<span style="background-color: #d6d6d6; padding-top: 0.2em; padding-bottom: 0.2em; padding-left: 0.5em; padding-right: 0.5em; border-radius: 2em; margin-left: 0.1em; margin-right: 0.1em; margin-top:0.01em; margin-bottom: 0.01em; color: #000; display: inline-block;">{{ prettyTag $key $val }}</span>
			{{ end }}
			</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal Cloudsweeper
</p>
`

const monthToDateTemplate = `
{{ $accountToUserMapping := .AccountToUser }}
<h2>Hello,</h2>

<p>
The following is a summary of this month's expenditures in {{ .CSP }}.
</p>
<p>
In the summary, only accounts with a total cost over ${{ .MinimumTotalCost }} are listed.
</p>
<p>
In the detailed breakdown, only costs over ${{ .MinimumCost }} are listed (but every cost is still counted towards the total!)
</p>

<h3>Summary:</h3>
{{ if gt (len .SortedUsers) 0 }}
	<table>
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Cost</strong></th>
		</tr>
	{{ range $i, $user := .SortedUsers }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ maybeRealName $user.Name $accountToUserMapping }}</td>
			<td>{{ printf "$%.2f" $user.TotalCost }}</td>
		</tr>
	{{ end }}
		<td colspan="2"><strong>Total cost: {{ printf "$%.2f" .TotalCost }}<strong></td>
	</table>
{{ end }}

<h3>Details:</h3>
{{ if gt (len .SortedUsers) 0 }}
	{{ range $index, $user := .SortedUsers }}
		<h3>{{- maybeRealName $user.Name $accountToUserMapping -}}'s costs:</h3>
		<h4>(Account ID: {{ $user.Name }})</h4>
		<table>
		<tr style="text-align:left;">
			<th><strong>Cost</strong></th>
			<th><strong>Description</strong></th>
		</tr>
		{{ range $i, $detailedCost := $user.DetailedCosts }}
			<tr {{ if even $i }} style="background-color: #f2f2f2;"{{ end }}>
				<td>{{ printf "$%.2f" $detailedCost.Cost }}</td>
				<td>{{ $detailedCost.Description }}</td>
			</tr>
		{{ end }}
		<td colspan="2"><strong>Total cost: {{ printf "$%.2f" $user.TotalCost }}<strong></td>
	</table>
	<br />
	{{ end }}
{{ end }}

<p>
Thank you,<br />
Your loyal Cloudsweeper
</p>
`
