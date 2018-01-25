package notify

const oldResourcesTemplate = `<h1>Hello {{ .Owner -}},</h1>

<p>
HouseKeeper has detected that you have some old resources still around. Please
take a look at them and clean them up if not needed.
</p>

<p>
<b>If you want to keep a resource and stop getting these reminders about it, add a tag with the key 
"whitelisted" to it.</b>
</p>

<p>
There is also an automated cleanup feature in HouseKeeper. To use this, please add
one of the following two types of tags (key: value) to your resource:
<br />
"housekeeper-lifetime: days-x", where x is the amount of days to keep the resource
<br />
"housekeeper-expiry: YYYY-MM-DD", to clean a resource up after the specified date, e.g. 2018-01-30
</p>

<h2>Old resources:</h2>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Instance type</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $instance.Owner }}</td>
			<td>{{ $instance.Location }}</td>
			<td>{{ $instance.ID }}</td>
			<td>{{ instname $instance }}</td>
			<td>{{ $instance.InstanceType }}</td>
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
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $image.Owner }}</td>
			<td>{{ $image.Location }}</td>
			<td>{{ $image.ID }}</td>
			<td>{{ $image.Name }}</td>
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
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Attached to instance</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Volume type</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $volume.Owner }}</td>
			<td>{{ $volume.Location }}</td>
			<td>{{ $volume.ID }}</td>
			<td>{{ $volume.SizeGB }} GB</td>
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
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $snapshot.Owner }}</td>
			<td>{{ $snapshot.Location }}</td>
			<td>{{ $snapshot.ID }}</td>
			<td>{{ $snapshot.SizeGB }} GB</td>
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
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Files</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Last modified</strong></th>
			<th><strong>Monthly cost</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $bucket.Owner }}</td>
			<td>{{ $bucket.ID }}</td>
			<td>{{ printf "%.3f GB" $bucket.TotalSizeGB }}</td>
			<td>{{ $bucket.ObjectCount }}</td>
			<td>{{ fdate $bucket.CreationTime "2006-01-02" }} ({{ daysrunning $bucket.CreationTime }})</td>
			<td>{{ fdate $bucket.LastModified "2006-01-02" }} ({{ daysrunning $bucket.LastModified }})</td>
			<td>{{ printf "$%.3f" (bucketcost $bucket) }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal housekeeper
</p>
`

const deletionWarningTemplate = `<h1>Hello {{ .Owner -}},</h1>

<h2>Resources will be cleaned up within {{ .Hours }} hours</h2>
<p>
Unless you take action, the resources listed below will be cleaned up 
from your account within the next {{ .Hours }} hours. <b>Make sure
you don't need to keep any of these resources</b>
</p>

<p>
If you want to save any of these resources, add a tag with the key <b>whitelisted</b>
</p>

<p>
Read more about how HouseKeeper works and how to better tag your resources at
<a href="#">this Wiki page</a>.
</p>

<h2>Old resources:</h2>
{{ if gt (len .Instances) 0 }}
	<h3>Instances</h3>
	<table style="width: 100%;">
		<tr style="text-align:left;">
			<th><strong>Account</strong></th>
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Instance type</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $instance := .Instances }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $instance.Owner }}</td>
			<td>{{ $instance.Location }}</td>
			<td>{{ $instance.ID }}</td>
			<td>{{ instname $instance }}</td>
			<td>{{ $instance.InstanceType }}</td>
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
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Name</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $image := .Images }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $image.Owner }}</td>
			<td>{{ $image.Location }}</td>
			<td>{{ $image.ID }}</td>
			<td>{{ $image.Name }}</td>
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
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Attached to instance</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Volume type</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $volume := .Volumes }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $volume.Owner }}</td>
			<td>{{ $volume.Location }}</td>
			<td>{{ $volume.ID }}</td>
			<td>{{ $volume.SizeGB }} GB</td>
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
			<th><strong>Location</strong></th>
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Total cost</strong></th>
		</tr>
	{{ range $i, $snapshot := .Snapshots }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $snapshot.Owner }}</td>
			<td>{{ $snapshot.Location }}</td>
			<td>{{ $snapshot.ID }}</td>
			<td>{{ $snapshot.SizeGB }} GB</td>
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
			<th><strong>ID</strong></th>
			<th><strong>Size (GB)</strong></th>
			<th><strong>Files</strong></th>
			<th><strong>Created</strong></th>
			<th><strong>Last modified</strong></th>
			<th><strong>Monthly cost</strong></th>
		</tr>
	{{ range $i, $bucket := .Buckets }}
		<tr {{ if even $i }}style="background-color: #f2f2f2;"{{ end }}>
			<td>{{ $bucket.Owner }}</td>
			<td>{{ $bucket.ID }}</td>
			<td>{{ printf "%.3f GB" $bucket.TotalSizeGB }}</td>
			<td>{{ $bucket.ObjectCount }}</td>
			<td>{{ fdate $bucket.CreationTime "2006-01-02" }} ({{ daysrunning $bucket.CreationTime }})</td>
			<td>{{ fdate $bucket.LastModified "2006-01-02" }} ({{ daysrunning $bucket.LastModified }})</td>
			<td>{{ printf "$%.3f" (bucketcost $bucket) }}</td>
		</tr>
	{{ end }}
	</table>
{{ end }}

<p>
Thank you,<br />
Your loyal housekeeper
</p>
`
